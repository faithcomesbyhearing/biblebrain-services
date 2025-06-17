-- name: GetOrganizations :many

SELECT 
    distinct bfco.organization_id,
    o.slug as organization_slug,
    ot.name as organization_name, ol.url as organization_logo_url
FROM bible_fileset_copyright_organizations as bfco
JOIN organizations o ON o.id = bfco.organization_id
INNER JOIN organization_translations ot ON ot.organization_id = o.id
LEFT JOIN organization_logos ol ON ol.organization_id = o.id AND ol.icon IS FALSE
WHERE bfco.organization_id IN (sqlc.slice('organizationsId'))
AND ot.language_id = 6414
ORDER BY organization_name;
