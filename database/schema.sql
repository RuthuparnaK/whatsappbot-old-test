
CREATE TABLE base_table (
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE user_account(
    id SERIAL PRIMARY KEY NOT NULL,
    name varchar(255),
    mobile_number bigint NOT NULL,
    last_ping TIMESTAMP NOT NULL,
    is_active boolean
 ) INHERITS (base_table);

 CREATE TABLE chat_history(
     id SERIAL PRIMARY KEY,
     message_received text,
     message_sent TEXT,
     message_id TEXT,
     message_status TEXT,
     sender_mobile_number BIGINT,
     sender_name TEXT,
     message_type TEXT,
     bunit_id BIGINT,
     app_name TEXT,
     brand_name TEXT
 ) INHERITS (base_table);

 CREATE TABLE msg_response(
     id SERIAL NOT NULL,
     request_msg text,
     response_msg text
 ) INHERITS (base_table);

 CREATE TABLE bunit_config(
     id SERIAL NOT NULL,
     bunit_id BIGINT NOT NULL,
     app_name TEXT NOT NULL,
     app_key TEXT NOT NULL,
     brand_name TEXT NOT NULL
 ) INHERITS (base_table);

 CREATE TABLE product_details (
	id SERIAL NOT NULL,
	name TEXT,
	product_url TEXT,
	short_url TEXT,
	bunit_id BIGINT
)