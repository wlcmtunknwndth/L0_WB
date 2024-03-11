package postgresql

const (
	getTemplate = `
	SELECT * 
	FROM orders 
	JOIN delivery ON orders.delivery = delivery.uid
	JOIN payment ON orders.order_uid = payment.transact
	JOIN items ON orders.track_number = items.track_number
	WHERE orders.order_uid = $1
	`

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
	$7
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
)
