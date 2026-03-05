// Copyright (C) 2026 Andrey Kriulin
// Licensed under the GNU Affero General Public License v3.0 or later.
// See the LICENSE file in the project root for the full license text.

package bcs

type orderDirection string

const (
	buy  orderDirection = "1"
	sell orderDirection = "2"
)

type orderType string

const (
	market orderType = "1"
	limit  orderType = "2"
)

type orderStatus string

const (
	New             orderStatus = "0"
	PartiallyFilled orderStatus = "1"
	Filled          orderStatus = "2"
	Cancelled       orderStatus = "4"
	Replaced        orderStatus = "5"
	Cancelling      orderStatus = "6"
	Rejected        orderStatus = "8"
	Replacing       orderStatus = "9"
	Pending         orderStatus = "10"
)

type recordStatus int32

const (
	RecordCanceld recordStatus = 1
	RecordDone    recordStatus = 2
	RecordActive  recordStatus = 3
)

type recordType int32

const (
	recordMarket             recordType = 1
	recordLimit              recordType = 2
	recordIceberg            recordType = 3
	recordStopLimit          recordType = 4
	recordTakeProfitLimit    recordType = 5
	recordStopLoss           recordType = 6
	recordTakeProfitStopLoss recordType = 7
	recordLimit30Days        recordType = 10
	recordTakeProfit         recordType = 11
	recordTrailingStop       recordType = 12
)

type recordDirection int32

const (
	recordBuy  = 1
	recordSell = 2
)
