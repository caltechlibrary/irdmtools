import csv, json

vocab = []
funder_mapping = {
    "AIR FORCE AP": "006gmme17",
    "AIR FORCE CR": "006gmme17",
    "AIR FORCE": "006gmme17",
    "ZZ DO NOT USE -AIR FORCE": "006gmme17",
    "Asian Office of Aerospace Research and Development": "011e9bt93",
    "DOD CR": "0447fe631",
    "DOD AP": "0447fe631",
    "ZZ - DO NOT USE - DEPARTMENT OF DEFENSE": "0447fe631",
    "ARMY CR": "00afsp483",
    "ARMY AP": "00afsp483",
    "ARMY": "00afsp483",
    "ZZ - DO NOT USE - ARMY": "00afsp483",
    "ONR AP": "00rk2pe57",
    "ONR CR 1": "00rk2pe57",
    "DARPA": "02caytj08",
    "DEFENSE ADVANCED RESEARCH PROJECT AGENCY": "02caytj08",
    "SNWS": "000ztjy10",
    "Naval Air Warfare Center Aircraft Division - Lakehurst": "03ar0mv07",
    "NAVAL COMMAND": "03ar0mv07",
    "NAVY": "03ar0mv07",
    "FLEET AND INDUSTRIAL SUPPLY CENTER": "03ar0mv07",
    "Naval Research Laboratory": "04d23a975",
    "SPACE AND NAVAL WARFARE SYSTEM": "000ztjy10",
    "Department of Homeland Security": "00jyr0d86",
    "Defense Threat Reduction Agency": "04tz64554",
    "NSA": "0047bvr32",
    "NATIONAL SECURITY AGENCY": "0047bvr32",
    "US Army Medical Research Command": "03cd02q50",
    "AA - DO NOT USE - US Army Medical Research Command": "03cd02q50",
    "NIST": "05xpvk416",
    "NOAA": "02z5nhe81",
    "USGS": "035a68863",
    "DOE CR": "01bj3aw27",
    "DOE LC": "01bj3aw27",
    "Department of Energy Pittsburgh": "01bj3aw27",
    "DEPARTMENT OF ENERGY, IL": "01bj3aw27",
    "Department of Energy Oak Ridge": "01bj3aw27",
    "DEPARTMENT OF ENERGY": "01bj3aw27",
    "Department of Energy Pittsburgh-ARRA": "01bj3aw27",
    "SANDIA NATIONAL LABORATORIES": "01apwpt12",
    "Federal Highway Administration": "0473rr271",
    "FAA": "05q0y0j38",
    "EPA": "03tns0030",
    "UNITED STATES ENVIRONMENTAL PROTECTION AGENCY": "03tns0030",
    "EPA LC": "03tns0030",
    "NASA Stennis": "027ka1x80",
    "NASA/Johnson Space Center": "027ka1x80",
    "NASA Ames": "027ka1x80",
    "ZZ - NASA HEADQUARTERS - DO NOT USE": "027ka1x80",
    "NASA Kennedy": "027ka1x80",
    "NASA": "027ka1x80",
    "NASA GLENN": "027ka1x80",
    "NASA LANGLEY": "027ka1x80",
    "NASA GODDARD LC": "027ka1x80",
    "NASA GODDARD CR": "027ka1x80",
    "NASA GODDARD": "027ka1x80",
    "NASA MARSHALL": "027ka1x80",
    "NASA NSSC": "027ka1x80",
    "NASA HOUSTON": "027ka1x80",
    "NASA HEADQUARTERS": "027ka1x80",
    "NASA AMES": "027ka1x80",
    "NASA SPECIAL": "027ka1x80",
    "NASA Johnson": "027ka1x80",
    "NASA WASHINGTON": "027ka1x80",
    "SMITHSONIAN": "01pp8nd67",
    "NSF": "021nxhr62",
    "NATIONAL SCIENCE FOUNDATION LIGO": "021nxhr62",
    "NATIONAL SCIENCE FOUNDATION ARRA": "021nxhr62",
    "NATIONAL SCIENCE FOUNDATION LIGO ARRA": "021nxhr62",
    "NATIONAL SCIENCE FOUNDATION": "021nxhr62",
    "National Science Foundation": "021nxhr62",
    "NSF-ARRA": "021nxhr62",
    "JPL": "027k65916",
    "NIH": "01cwqze88",
    "NATIONAL INSTITUTES OF HEALTH": "01cwqze88",
    "NATIONAL INSTITUTES OF HEALTH ARRA": "01cwqze88",
    "ASPR/BARDA": "029y69023",
    "NIH LC": "01cwqze88",
    "NIH CR": "01cwqze88",
    "USAID": "01n6e6j62",
    "Homeland Security Advanced Research Projects Agency": "00jyr0d86",
    "Homeland Security Adv Research Projects Agency": "00jyr0d86",
    "Centers for Disease Control and Prevention": "042twtr12",
    "USDA": "01na82s61",
    "UNITED STATES DEPARTMENT OF AGRICULTURE": "01na82s61",
    "National Historical Publications and Records Commission": "032214n64",
    "ZZ - DO NOT USE NATIONAL ENDOWMENT FOR THE HUMANITIES": "02vdm1p28",
    "Department of State": "03vvynj75",
    "U.S. Naval Observatory": "048s2rn92",
    "Department of Justice": "02916qm60",
    "FEMA": "01g9x3v85",
    "Microelectronics Advanced Research Corporation": "047z4n946",
    "USNRC": "03nhmbj89",
    "FOOD AND DRUG ADMINISTRATION": "034xvzb47",
    "Photonic Systems, Inc.": "016s82z56",
    "National Geospatial-Intelligence Agency": "02k4pxv54",
    "Department of the Interior": "03v0pmy70",
    "NATIONAL INSTITUTE OF STANDARDS AND TECHNOLOGY": "05xpvk416",
    "NATIONAL OCEANIC AND ATMOSPHERIC ADMINISTRATION": "02z5nhe81",
    "DEPARTMENT OF EDUCATION": "05nne8c43",
    "UNITED STATES GEOLOGICAL SURVEY": "035a68863",
    "United States Geological Survey-ARRA": "035a68863",
    "United States Geological Survey": "035a68863",
    "US GEOLOGICAL SURVEY": "035a68863",
    "United States Bureau of Reclamation": "00ezrrm21",
}
with open("awards.csv", "r") as f:
    reader = csv.DictReader(f)
    awards = {}
    for award in reader:
        awards[award["Award #"]] = award
    for award in awards:
        data = awards[award]
        title = data["Award Full Name"]
        if data["Prime Funding Source"] != "":
            try:
                funder = funder_mapping[data["Prime Funding Source"]]
            except:
                print("No mapping for", data["Prime Funding Source"])
                exit()
            number = data["Prime Agreement #"]
        else:
            try:
                funder = funder_mapping[data["Oracle Funding Source Name"]]
            except:
                print("No mapping for", data["Oracle Funding Source Name"])
                exit()
            number = data["Funding Src Award #"]
        vocab.append(
            {
                "id": award,
                "title": {"en": title},
                "number": number,
                "funder": {"id": funder},
            }
        )

with open("awards.jsonl", "w") as f:
    for award in vocab:
        f.write(json.dumps(award) + "\n")
