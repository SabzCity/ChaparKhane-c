/* For license and copyright information please see LEGAL file in repository */

package datastore

import "testing"

func TestProductAuction_CalculatePayablePrice(t *testing.T) {
	tests := []struct {
		name string
		pa   *ProductAuction
	}{
		{
			name: "test1",
			pa: &ProductAuction{
				SuggestPrice: 1000,
				Discount:     1400,
				PayablePrice: 860,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expected = tt.pa.PayablePrice
			tt.pa.CalculatePayablePrice()
			if tt.pa.PayablePrice != expected {
				t.Errorf("Failed on = %v , Need %v got %v", tt.name, expected, tt.pa.PayablePrice)
			}
		})
	}
}
