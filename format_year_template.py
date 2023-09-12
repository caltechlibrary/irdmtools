with open("year_template_formatted.html", "w") as w:
    with open("year_template.html", "r") as f:
        template = f.readlines()
        for line in template:
            if "href=" in line:
                year = line.split('href="')[1].split(".html")[0]
                url = f"/search?q=metadata.publication_date%3A%5B{year}-01%20TO%20{year}-12-31%5D"
                w.write(line.replace(f"{year}.html", url))
            else:
                w.write(line)
