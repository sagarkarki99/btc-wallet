CREATE OR REPLACE FUNCTION get_wallet_info (walletId integer)
RETURNS TABLE (
    id INTEGER,
    xpub TEXT,
    accountid INTEGER,
    next_index INTEGER
)
LANGUAGE sql STABLE
 AS $fun$
    SELECT w.id, w.xpub, w.accountid, a.next_index
    FROM wallet w
    JOIN address a
     ON w.accountid = a.account_id
    WHERE w.accountid = walletId
    ORDER BY a.next_index DESC
    LIMIT 1
$fun$ ;