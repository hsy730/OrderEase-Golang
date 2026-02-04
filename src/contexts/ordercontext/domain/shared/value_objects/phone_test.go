package value_objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid mobile number - 138",
			phone:   "13812345678",
			wantErr: false,
		},
		{
			name:    "valid mobile number - 139",
			phone:   "13912345678",
			wantErr: false,
		},
		{
			name:    "valid mobile number - 150",
			phone:   "15012345678",
			wantErr: false,
		},
		{
			name:    "valid mobile number - 186",
			phone:   "18612345678",
			wantErr: false,
		},
		{
			name:    "valid mobile number - 177",
			phone:   "17712345678",
			wantErr: false,
		},
		{
			name:    "valid mobile number - 199",
			phone:   "19912345678",
			wantErr: false,
		},
		{
			name:    "empty string - allowed",
			phone:   "",
			wantErr: false,
		},
		{
			name:    "invalid - too short",
			phone:   "138123456",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
		{
			name:    "invalid - too long",
			phone:   "138123456789",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
		{
			name:    "invalid - not starting with 1",
			phone:   "23812345678",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
		{
			name:    "invalid - starting with 10",
			phone:   "10123456789",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
		{
			name:    "invalid - contains letters",
			phone:   "138123456ab",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
		{
			name:    "invalid - contains special chars",
			phone:   "1381234567-",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
		{
			name:    "invalid - with spaces",
			phone:   "138 1234 5678",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
		{
			name:    "invalid - chinese characters",
			phone:   "一二三四五六七八九零",
			wantErr: true,
			errMsg:  "手机号必须为11位数字且以1开头",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhone(tt.phone)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, Phone(""), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, Phone(tt.phone), got)
			}
		})
	}
}

func TestPhone_String(t *testing.T) {
	phone := Phone("13812345678")
	assert.Equal(t, "13812345678", phone.String())
}

func TestPhone_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		phone    Phone
		wantValid bool
	}{
		{
			name:      "valid phone number",
			phone:     Phone("13812345678"),
			wantValid: true,
		},
		{
			name:      "empty phone - considered valid",
			phone:     Phone(""),
			wantValid: true,
		},
		{
			name:      "invalid phone - too short",
			phone:     Phone("138123456"),
			wantValid: false,
		},
		{
			name:      "invalid phone - starts with 2",
			phone:     Phone("23812345678"),
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.phone.IsValid()
			assert.Equal(t, tt.wantValid, got)
		})
	}
}

func TestPhone_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		phone    Phone
		wantEmpty bool
	}{
		{
			name:      "empty phone",
			phone:     Phone(""),
			wantEmpty: true,
		},
		{
			name:      "non-empty phone",
			phone:     Phone("13812345678"),
			wantEmpty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.phone.IsEmpty()
			assert.Equal(t, tt.wantEmpty, got)
		})
	}
}

func TestPhone_Masked(t *testing.T) {
	tests := []struct {
		name         string
		phone        Phone
		wantMasked   string
	}{
		{
			name:       "normal phone number",
			phone:      Phone("13812345678"),
			wantMasked: "138****5678",
		},
		{
			name:       "another phone number",
			phone:      Phone("18698765432"),
			wantMasked: "186****5432",
		},
		{
			name:       "empty phone",
			phone:      Phone(""),
			wantMasked: "",
		},
		{
			name:       "invalid length - too short",
			phone:      Phone("138123456"),
			wantMasked: "138123456",
		},
		{
			name:       "invalid length - too long",
			phone:      Phone("138123456789"),
			wantMasked: "138123456789",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.phone.Masked()
			assert.Equal(t, tt.wantMasked, got)
		})
	}
}

func TestPhone_Carrier(t *testing.T) {
	tests := []struct {
		name        string
		phone       Phone
		wantCarrier string
	}{
		// 移动号段
		{
			name:        "China Mobile - 134",
			phone:       Phone("13412345678"),
			wantCarrier: "移动",
		},
		{
			name:        "China Mobile - 138",
			phone:       Phone("13812345678"),
			wantCarrier: "移动",
		},
		{
			name:        "China Mobile - 139",
			phone:       Phone("13912345678"),
			wantCarrier: "移动",
		},
		{
			name:        "China Mobile - 150",
			phone:       Phone("15012345678"),
			wantCarrier: "移动",
		},
		{
			name:        "China Mobile - 182",
			phone:       Phone("18212345678"),
			wantCarrier: "移动",
		},
		{
			name:        "China Mobile - 188",
			phone:       Phone("18812345678"),
			wantCarrier: "移动",
		},
		// 联通号段
		{
			name:        "China Unicom - 130",
			phone:       Phone("13012345678"),
			wantCarrier: "联通",
		},
		{
			name:        "China Unicom - 131",
			phone:       Phone("13112345678"),
			wantCarrier: "联通",
		},
		{
			name:        "China Unicom - 132",
			phone:       Phone("13212345678"),
			wantCarrier: "联通",
		},
		{
			name:        "China Unicom - 185",
			phone:       Phone("18512345678"),
			wantCarrier: "联通",
		},
		{
			name:        "China Unicom - 186",
			phone:       Phone("18612345678"),
			wantCarrier: "联通",
		},
		// 电信号段
		{
			name:        "China Telecom - 133",
			phone:       Phone("13312345678"),
			wantCarrier: "电信",
		},
		{
			name:        "China Telecom - 180",
			phone:       Phone("18012345678"),
			wantCarrier: "电信",
		},
		{
			name:        "China Telecom - 189",
			phone:       Phone("18912345678"),
			wantCarrier: "电信",
		},
		{
			name:        "China Telecom - 199",
			phone:       Phone("19912345678"),
			wantCarrier: "电信",
		},
		// 未知号段
		{
			name:        "Unknown carrier - 120",
			phone:       Phone("12012345678"),
			wantCarrier: "未知",
		},
		{
			name:        "empty phone",
			phone:       Phone(""),
			wantCarrier: "未知",
		},
		{
			name:        "invalid length",
			phone:       Phone("138123456"),
			wantCarrier: "未知",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.phone.Carrier()
			assert.Equal(t, tt.wantCarrier, got)
		})
	}
}
