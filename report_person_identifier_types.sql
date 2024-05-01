--
-- Report out how many of each identifiers we find in creators
--
select 
  count(*) as type_count,
  jsonb_path_query(json, '$.metadata.creators[*].person_or_org.identifiers[*].scheme')->>0 as identifier_type
from rdm_records_metadata as T
group by T.identifier_type ;
