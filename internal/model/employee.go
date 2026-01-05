package model

import (
	"time"
)

type Employee struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OutletID  uint      `json:"outlet_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `gorm:"unique" json:"email"`
	Password  string    `json:"-"` // Password disembunyikan dari JSON
	Role      string    `json:"role"` // Owner / Pegawai
	
	// Permissions (Boolean Columns)
	PermMakeOrder         bool `json:"perm_make_order"`
	PermCancelOrder       bool `json:"perm_cancel_order"`
	PermManageExpenses    bool `json:"perm_manage_expenses"`
	PermManageServices    bool `json:"perm_manage_services"`
	PermViewRevenue       bool `json:"perm_view_revenue"`
	PermManageEmployees   bool `json:"perm_manage_employees"`
	PermReportTransaction bool `json:"perm_report_transaction"`
	PermReportPerformance bool `json:"perm_report_performance"`
	PermReportFinance     bool `json:"perm_report_finance"`
	PermReportCustomer    bool `json:"perm_report_customer"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}