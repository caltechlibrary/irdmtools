if [ -f test-ids.txt ]; then
	rm test-ids.txt
fi
touch test-ids.txt

#for TYPE in article book book_section  conference_item dataset monograph other patent software teaching_resource thesis video website; do
for TYPE in other website; do
	mysql caltechauthors -e "SELECT eprintid FROM eprint WHERE type = '${TYPE}' AND eprint_status = 'archive' ORDER BY RAND() LIMIT 25" | grep -v 'eprintid' >>test-ids.txt 
done
