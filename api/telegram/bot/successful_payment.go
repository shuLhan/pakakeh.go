// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// SuccessfulPayment contains basic information about a successful payment.
type SuccessfulPayment struct {
	// Optional. Order info provided by the user.
	OrderInfo *OrderInfo `json:"order_info"`

	// Three-letter ISO 4217 currency code.
	Currency string `json:"currency"`

	// Bot specified invoice payload.
	InvoicePayload string `json:"invoice_payload"`

	// Optional. Identifier of the shipping option chosen by the user.
	ShippingOptionID string `json:"shipping_option_id"`

	// Telegram payment identifier.
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`

	// Provider payment identifier.
	ProviderPaymentChargeID string `json:"provider_payment_charge_id"`

	// Total price in the smallest units of the currency (integer, not
	// float/double). For example, for a price of US$ 1.45 pass amount =
	// 145. See the exp parameter in currencies.json, it shows the number
	// of digits past the decimal point for each currency (2 for the
	// majority of currencies).
	TotalAmount int `json:"total_amount"`
}
