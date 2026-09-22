package monitor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wabisabi926/faststrm/internal/model"
	"github.com/wabisabi926/faststrm/internal/service/client115"
	"github.com/wabisabi926/faststrm/internal/service/db"
)

// TestFindLocalStrmByFileName_RootDirAndSubDir 测试本地同名 STRM 查找（根目录及子目录）
func TestFindLocalStrmByFileName_RootDirAndSubDir(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. 在根目录放置一个 STRM 文件
	rootStrm := filepath.Join(tmpDir, "Ballerina.2025.strm")
	if err := os.WriteFile(rootStrm, []byte("http://test/1"), 0o644); err != nil {
		t.Fatalf("write rootStrm: %v", err)
	}

	// 2. 在子目录放置一个 STRM 文件
	subDir := filepath.Join(tmpDir, "SubDir", "Season 1")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir subDir: %v", err)
	}
	nestedStrm := filepath.Join(subDir, "Episode1.strm")
	if err := os.WriteFile(nestedStrm, []byte("http://test/2"), 0o644); err != nil {
		t.Fatalf("write nestedStrm: %v", err)
	}

	mappings := []model.MonitorPathMapping{
		{Account: "test-acc", CloudPath: "movie", LocalPath: tmpDir},
	}
	m := &Monitor{}

	// 测试根目录文件查找
	foundRoot := m.findLocalStrmByFileName("test-acc", "Ballerina.2025.mkv", 1, mappings)
	if foundRoot != rootStrm {
		t.Errorf("expected %q, got %q", rootStrm, foundRoot)
	}

	// 测试子目录嵌套文件查找
	foundNested := m.findLocalStrmByFileName("test-acc", "Episode1.mkv", 1, mappings)
	if foundNested != nestedStrm {
		t.Errorf("expected %q, got %q", nestedStrm, foundNested)
	}

	// 测试不存在的文件
	foundNonExistent := m.findLocalStrmByFileName("test-acc", "NonExistent.mkv", 1, mappings)
	if foundNonExistent != "" {
		t.Errorf("expected empty string, got %q", foundNonExistent)
	}
}

// TestHandleDeleteFallback_Success 测试删除事件未命中 mapping 时的本地兜底删除
func TestHandleDeleteFallback_Success(t *testing.T) {
	tmpDir := t.TempDir()
	strmPath := filepath.Join(tmpDir, "Ballerina.2025.strm")
	relatedNfo := filepath.Join(tmpDir, "Ballerina.2025.nfo")
	if err := os.WriteFile(strmPath, []byte("http://strm"), 0o644); err != nil {
		t.Fatalf("write strmPath: %v", err)
	}
	if err := os.WriteFile(relatedNfo, []byte("<nfo>"), 0o644); err != nil {
		t.Fatalf("write relatedNfo: %v", err)
	}

	// 初始化 sqlite
	sqliteDB, err := db.OpenNew(tmpDir)
	if err != nil {
		t.Fatalf("db.OpenNew: %v", err)
	}
	defer sqliteDB.Close()

	// 预先插入一条记录
	_ = db.UpsertFilePathEntry(sqliteDB, "test-acc", db.FilePathEntry{
		FileID:   "12345",
		Path:     "movie/Ballerina.2025.mkv",
		FileName: "Ballerina.2025.mkv",
		PickCode: "abcdefghij1234567",
	})

	cfg := model.LifeMonitorSettings{
		Enabled: true,
		EventTypes: model.EventTypesSettings{
			Remove: true,
		},
		PathMappings: []model.MonitorPathMapping{
			{Account: "test-acc", CloudPath: "movie", LocalPath: tmpDir},
		},
	}
	m := &Monitor{
		settingsFn: func() model.LifeMonitorSettings { return cfg },
		sqliteDB:   sqliteDB,
	}

	event := client115.LifeEventItem{
		FileID:       "12345",
		FileName:     "Ballerina.2025.mkv",
		FileCategory: 1,
		Type:         22,
	}

	if err := m.handleDeleteFallback(context.Background(), "test-acc", event, "Ballerina.2025.mkv"); err != nil {
		t.Fatalf("handleDeleteFallback: %v", err)
	}

	// 验证 STRM 及关联文件已被删除
	if _, err := os.Stat(strmPath); !os.IsNotExist(err) {
		t.Errorf("strm 文件应被删除，但仍然存在")
	}
	if _, err := os.Stat(relatedNfo); !os.IsNotExist(err) {
		t.Errorf("关联 nfo 文件应被删除，但仍然存在")
	}

	// 验证 DB 记录已被清理
	entry, err := db.GetFilePathEntry(sqliteDB, "test-acc", "12345")
	if err != nil {
		t.Errorf("GetFilePathEntry err: %v", err)
	}
	if entry != nil {
		t.Errorf("DB 中的 entry 应被清理，但得到: %+v", entry)
	}
}

// TestWriteAheadDB_PreventSingleSegUnmapped 测试 Write-Ahead DB 防止裸文件名污染
func TestWriteAheadDB_PreventSingleSegUnmapped(t *testing.T) {
	tmpDir := t.TempDir()
	sqliteDB, err := db.OpenNew(tmpDir)
	if err != nil {
		t.Fatalf("db.OpenNew: %v", err)
	}
	defer sqliteDB.Close()

	cfg := model.LifeMonitorSettings{
		Enabled: true,
		EventTypes: model.EventTypesSettings{
			Remove: true,
		},
		PathMappings: []model.MonitorPathMapping{
			{Account: "test-acc", CloudPath: "movie", LocalPath: tmpDir},
		},
	}
	m := &Monitor{
		settingsFn: func() model.LifeMonitorSettings { return cfg },
		sqliteDB:   sqliteDB,
	}

	event := client115.LifeEventItem{
		FileID:       "99999",
		FileName:     "DirtyFile.mkv",
		FileCategory: 1,
		Type:         22,
	}

	// 传入裸文件名且未命中映射
	decision, _ := m.preProcessEventWithSource(context.Background(), "test-acc", event, "DirtyFile.mkv", "BARE_FILENAME", cfg)
	if decision.SkipReason != "no_path_mapping" {
		t.Errorf("expected skipReason 'no_path_mapping', got %q", decision.SkipReason)
	}

	// 验证 DB 中并没有写入 99999
	entry, err := db.GetFilePathEntry(sqliteDB, "test-acc", "99999")
	if err != nil {
		t.Errorf("GetFilePathEntry: %v", err)
	}
	if entry != nil {
		t.Errorf("未命中的裸文件名不应写入 DB，但查到: %+v", entry)
	}
}
