package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // pure-Go SQLite driver
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: dbstat <system.db路径>")
		os.Exit(1)
	}
	dbPath, _ := filepath.Abs(os.Args[1])
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("打开数据库失败:", err)
		os.Exit(1)
	}
	defer db.Close()

	stats := []struct{ name, sql string }{
		{"sys_users", "SELECT count(*) FROM sys_users WHERE deleted_at IS NULL"},
		{"sys_departments", "SELECT count(*) FROM sys_departments WHERE deleted_at IS NULL"},
		{"sys_positions", "SELECT count(*) FROM sys_positions WHERE deleted_at IS NULL"},
		{"pm_entities(eps_node)", "SELECT count(*) FROM pm_entities WHERE entity_type='eps_node' AND deleted_at IS NULL"},
		{"pm_entities(task)", "SELECT count(*) FROM pm_entities WHERE entity_type='task' AND deleted_at IS NULL"},
		{"pm_entities(team_member)", "SELECT count(*) FROM pm_entities WHERE entity_type='team_member' AND deleted_at IS NULL"},
		{"pm_entities(milestone)", "SELECT count(*) FROM pm_entities WHERE entity_type='milestone' AND deleted_at IS NULL"},
		{"pm_entities(cost_item)", "SELECT count(*) FROM pm_entities WHERE entity_type='cost_item' AND deleted_at IS NULL"},
		{"pm_entities(requirement)", "SELECT count(*) FROM pm_entities WHERE entity_type='requirement' AND deleted_at IS NULL"},
		{"pm_entities(issue)", "SELECT count(*) FROM pm_entities WHERE entity_type='issue' AND deleted_at IS NULL"},
		{"pm_entities(risk)", "SELECT count(*) FROM pm_entities WHERE entity_type='risk' AND deleted_at IS NULL"},
		{"pm_entities(change_request)", "SELECT count(*) FROM pm_entities WHERE entity_type='change_request' AND deleted_at IS NULL"},
		{"pm_entities(deliverable)", "SELECT count(*) FROM pm_entities WHERE entity_type='deliverable' AND deleted_at IS NULL"},
		{"pm_entities(scope_item)", "SELECT count(*) FROM pm_entities WHERE entity_type='scope_item' AND deleted_at IS NULL"},
		{"pm_entities(status=archived)", "SELECT count(*) FROM pm_entities WHERE status='archived' AND deleted_at IS NULL"},
		{"pm_time_entries", "SELECT count(*) FROM pm_time_entries WHERE deleted_at IS NULL"},
		{"pm_cost_actuals", "SELECT count(*) FROM pm_cost_actuals WHERE deleted_at IS NULL"},
		{"pm_baselines", "SELECT count(*) FROM pm_baselines WHERE deleted_at IS NULL"},
		{"pm_workflow_instances", "SELECT count(*) FROM pm_workflow_instances"},
		{"pm_approval_records", "SELECT count(*) FROM pm_approval_records WHERE deleted_at IS NULL"},
		{"pm_relations", "SELECT count(*) FROM pm_relations"},
		{"pm_report_snapshots", "SELECT count(*) FROM pm_report_snapshots WHERE deleted_at IS NULL"},
		{"pm_task_links", "SELECT count(*) FROM pm_task_links"},
	}
	fmt.Println("=== 数据库统计 ===")
	for _, s := range stats {
		var cnt int64
		if err := db.QueryRow(s.sql).Scan(&cnt); err != nil {
			fmt.Printf("  %-35s ERR: %v\n", s.name, err)
			continue
		}
		fmt.Printf("  %-35s %d\n", s.name, cnt)
	}
}
