-- Verificar si existe el source El Observatodo
SELECT 
    id, 
    name, 
    url, 
    type, 
    fetch_strategy, 
    interval_minutes, 
    last_run_at, 
    created_at,
    config
FROM sources 
WHERE url LIKE '%observatodo%' 
   OR name LIKE '%Observatodo%'
   OR url LIKE '%elobservatodo%';

-- Verificar todos los sources activos
SELECT 
    COUNT(*) as total_sources,
    SUM(CASE WHEN url LIKE '%observatodo%' THEN 1 ELSE 0 END) as observatodo_sources,
    SUM(CASE WHEN url LIKE '%diarioeldia%' THEN 1 ELSE 0 END) as diarioeldia_sources
FROM sources 
WHERE url IS NOT NULL;

-- Verificar sources que deberían ejecutarse ahora
SELECT 
    id,
    name,
    url,
    interval_minutes,
    last_run_at,
    CASE 
        WHEN last_run_at IS NULL THEN 'NUNCA EJECUTADO'
        WHEN last_run_at + (interval_minutes || ' minutes')::interval < NOW() THEN 'DEBERÍA EJECUTARSE'
        ELSE 'NO DEBERÍA EJECUTARSE TODAVÍA'
    END as status,
    NOW() as current_time,
    last_run_at + (interval_minutes || ' minutes')::interval as next_run_time
FROM sources 
WHERE url IS NOT NULL 
AND (last_run_at IS NULL OR last_run_at + (interval_minutes || ' minutes')::interval < NOW())
ORDER BY last_run_at NULLS FIRST
LIMIT 10;