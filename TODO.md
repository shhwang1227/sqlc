 insert into  authors (id,name,bio,company_id) values(1,'a','a',1),(2,'b','b',2) ON DUPLICATE KEY UPDATE name='c' ,bio='d';
 这种标准的upsert 不支持，只支持特殊语法

 insert into  authors (id,name,bio,company_id) values(1,'a','a',1),(2,'b','b',2) ON DUPLICATE KEY UPDATE name=values(`name`) ,bio=values(`bio`);


 /*  name: Companys :execresult */
select * from company wehre id > ? and id < ?;

/*  name: Companys :execresult */
select * from company wehre id between ? and ?;