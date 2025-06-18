-- name: GetFilesetCopyrights :many
SELECT 
    GROUP_CONCAT(DISTINCT bfco.organization_id) AS organization_id_list, bible_fileset_copyrights.copyright_date,
    bible_fileset_copyrights.copyright,
    bft.description product_code
FROM bible_fileset_copyrights
JOIN (SELECT bfco2.hash_id, bfco2.organization_id FROM bible_fileset_copyright_organizations bfco2 GROUP BY bfco2.hash_id, bfco2.organization_id) bfco ON bfco.hash_id = bible_fileset_copyrights.hash_id
JOIN bible_fileset_tags bft ON bft.hash_id = bible_fileset_copyrights.hash_id AND bft.name = 'stock_no'
WHERE bft.description IN (sqlc.slice('productCodes'))
AND EXISTS (
    SELECT 1
    FROM bible_filesets bf
    WHERE bf.hash_id = bible_fileset_copyrights.hash_id
    AND bf.set_type_code IN (sqlc.slice('typeCodes'))
)
GROUP BY
    bible_fileset_copyrights.copyright_date,
    bible_fileset_copyrights.copyright,
    bft.description
ORDER BY product_code;
