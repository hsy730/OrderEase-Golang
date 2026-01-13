package shop

import (
	"errors"
	"time"

	"orderease/domain/order"
	"orderease/domain/shared"
)

type Shop struct {
	ID            shared.ID
	Name          string
	OwnerUsername string
	OwnerPassword string
	ContactPhone  string
	ContactEmail  string
	Address       string
	ImageURL      string
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ValidUntil    time.Time
	Settings      string
	OrderStatusFlow order.OrderStatusFlow
}

func NewShop(name, ownerUsername, ownerPassword string, validUntil time.Time) (*Shop, error) {
	if name == "" {
		return nil, errors.New("店铺名称不能为空")
	}

	if ownerUsername == "" {
		return nil, errors.New("店主用户名不能为空")
	}

	if ownerPassword == "" {
		return nil, errors.New("店主密码不能为空")
	}

	if validUntil.IsZero() {
		validUntil = time.Now().AddDate(1, 0, 0)
	}

	now := time.Now()

	return &Shop{
		Name:          name,
		OwnerUsername: ownerUsername,
		OwnerPassword: ownerPassword,
		ValidUntil:    validUntil,
		CreatedAt:     now,
		UpdatedAt:     now,
		OrderStatusFlow: order.OrderStatusFlow{
			Statuses: getDefaultOrderStatuses(),
		},
	}, nil
}

func (s *Shop) IsExpired() bool {
	now := time.Now().UTC()
	return s.ValidUntil.Before(now)
}

func (s *Shop) RemainingDays() int {
	hours := time.Until(s.ValidUntil.UTC()).Hours()
	return int(hours / 24)
}

func (s *Shop) UpdateBasicInfo(name, contactPhone, contactEmail, address, description string) error {
	if name != "" {
		s.Name = name
	}
	if contactPhone != "" {
		s.ContactPhone = contactPhone
	}
	if contactEmail != "" {
		s.ContactEmail = contactEmail
	}
	if address != "" {
		s.Address = address
	}
	if description != "" {
		s.Description = description
	}
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Shop) UpdateOrderStatusFlow(flow order.OrderStatusFlow) error {
	if len(flow.Statuses) == 0 {
		return errors.New("订单流转配置不能为空")
	}
	s.OrderStatusFlow = flow
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Shop) UpdateValidUntil(validUntil time.Time) error {
	if validUntil.IsZero() {
		return errors.New("有效期不能为空")
	}
	s.ValidUntil = validUntil
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Shop) UpdatePassword(newPassword string) error {
	if newPassword == "" {
		return errors.New("新密码不能为空")
	}
	s.OwnerPassword = newPassword
	s.UpdatedAt = time.Now()
	return nil
}

func getDefaultOrderStatuses() []order.OrderStatusConfig {
	return []order.OrderStatusConfig{
		{
			Value:   0,
			Label:   "待处理",
			Type:    "warning",
			IsFinal: false,
			Actions: []order.OrderStatusTransition{
				{Name: "接单", NextStatus: 1, NextStatusLabel: "已接单"},
				{Name: "取消", NextStatus: 10, NextStatusLabel: "已取消"},
			},
		},
		{
			Value:   1,
			Label:   "已接单",
			Type:    "primary",
			IsFinal: false,
			Actions: []order.OrderStatusTransition{
				{Name: "完成", NextStatus: 9, NextStatusLabel: "已完成"},
				{Name: "取消", NextStatus: 10, NextStatusLabel: "已取消"},
			},
		},
		{
			Value:   9,
			Label:   "已完成",
			Type:    "success",
			IsFinal: true,
			Actions: []order.OrderStatusTransition{},
		},
		{
			Value:   10,
			Label:   "已取消",
			Type:    "info",
			IsFinal: true,
			Actions: []order.OrderStatusTransition{},
		},
	}
}
