package domain

import (
	ordersv1 "avtoplaneta/gen/orders/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto конвертирует модель OrderItem в Protobuf OrderItem.
func (item *OrderItem) ToProto() *ordersv1.OrderItem {
	if item == nil {
		return nil
	}
	return &ordersv1.OrderItem{
		Id:       item.ID,
		OrderId:  item.OrderID,
		PartId:   item.PartID,
		Quantity: int32(item.Quantity),
		Price:    item.Price,
	}
}

// ToProto конвертирует модель Order в Protobuf Order.
func (o *Order) ToProto() *ordersv1.Order {
	if o == nil {
		return nil
	}
	items := make([]*ordersv1.OrderItem, len(o.Items))
	for i := range o.Items {
		items[i] = o.Items[i].ToProto()
	}

	return &ordersv1.Order{
		Id:                 o.ID,
		CustomerId:         o.CustomerID,
		SellerId:           o.SellerID,
		Seller:             o.Seller,
		Part:               o.Part,
		PartId:             o.PartID,
		Location:           o.Location,
		BuyerNumber:        o.BuyerNumber,
		Status:             o.Status,
		StatusText:         o.StatusText,
		AutoDeleted:        o.AutoDeleted,
		CreatedAt:          timestamppb.New(o.CreatedAt),
		CreatedAtFormatted: o.CreatedAtFormatted,
		TimeAgo:            o.TimeAgo,
		Items:              items,
	}
}
