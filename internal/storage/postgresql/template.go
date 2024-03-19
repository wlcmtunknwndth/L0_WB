package postgresql

const (
	getOrderTemplate = `
SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
	o.customer_id, o.delivery_service, o.shardkey, o.sm_id, 
		o.date_created, o.oof_shard, 
	d.fio, d.phone, d.zip, d.city, d.address, d.region, d.email,
	pa.request_id, pa.currency, pa.provider, pa.amount, pa.payment_dt, pa.bank,
		pa.delivery_cost, pa.goods_total, pa.custom_fee
FROM orders o  
JOIN delivery d ON o.track_number = d.track_number
JOIN payment pa ON o.order_uid = pa.transact
WHERE o.order_uid = $1
`
	getItemsTemplate = `SELECT * FROM items WHERE track_number = $1`

	saveOrder = `
INSERT INTO orders(
	order_uid,
	track_number,
	entry,
	locale,
	internal_signature,
	customer_id,
	delivery_service,
	shardkey,
	sm_id,
	date_created,
	oof_shard
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11
);`

	saveDelivery = `
INSERT INTO delivery(
	track_number,
    fio,
	phone,
	zip,
	city,
	address,
	region,
	email
) 
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
    $8
);
`

	savePayment = `
INSERT INTO payment VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10
);
`

	saveItems = `
INSERT INTO items VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11
);
`

	deletePayment = `
DELETE FROM payment WHERE transact = $1;
`
	deleteItems = ` 
DELETE FROM items WHERE track_number = $1;
`
	deleteDelivery = `
DELETE FROM delivery WHERE track_number = $1;
`
	deleteOrder = `
DELETE FROM orders WHERE order_uid = $1;
`

	deleteCache = `
DELETE FROM cached WHERE order_uid = $1
`
	saveCache       = `INSERT INTO cached VALUES($1)`
	getCache        = `SELECT * FROM cached`
	isAlreadyCached = `SELECT * FROM cached WHERE order_uid = $1`
)
