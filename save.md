```sql
WITH inserted_tag AS (
  INSERT INTO tags (tag_name, account_id)
  SELECT unnest(t.tag_names), $3
  FROM (SELECT $2::text[] AS tag_names) t
  ON CONFLICT (tag_name) DO NOTHING
  RETURNING tag_id, tag_name
), existing_tag AS (
  SELECT tag_id, tag_name
  FROM tags
  WHERE account_id = $3 AND tag_name = ANY($2)
)
INSERT INTO link_tags (link_id, tag_id)
SELECT l.link_id, COALESCE(it.tag_id, et.tag_id)
FROM unnest($1::uuid[]) l(link_id)
CROSS JOIN (
  SELECT COALESCE(it.tag_id, et.tag_id) as tag_id
  FROM inserted_tag it
  FULL JOIN existing_tag et ON it.tag_name = et.tag_name
) t
RETURNING *;
```
