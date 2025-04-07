begin;


alter table payments add column refund_id varchar(255);
UPDATE payments SET status = 'SETTLED' WHERE status = 'CLEARED'; -- Check this app side also
-- Add webhook refund response

commit;
