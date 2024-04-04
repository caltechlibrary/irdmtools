package irdmtools

import (
	//"os"
	"testing"
)

func TestDataCiteToRDM(t *testing.T) {
	//FIXME: Need an arxiv DOI to look up at DataCite
	arxiv_ids := []string{
		"10.22002/D1.868",
		"arXiv:2312.07215",
		"arXiv:2305.06519",
		"arXiv:2312.03791",
		"arXiv:2305.19279",
		"arXiv:2305.05315",
		"arXiv:2305.07673",
		"arXiv:2111.03606",
		"arXiv:2112.06016",
		/* these arXiv ids don't seem to have data at DataCite ...  */
		//"arXiv:2402.12335v1",
		//"arXiv:2401.12460v1",
		//"arXiv:2204.13532v2",
	}
	app := new(Doi2Rdm)
	app.Cfg = new(Config)
	for _, doi := range arxiv_ids {
		options := new(Doi2RdmOptions)
		options.MailTo = "dld-test@library.caltech.edu"
		obj, err := QueryDataCiteObject(app.Cfg, doi, options)
		if err != nil {
			t.Error(err)
		}
		if obj == nil {
			t.Errorf("expected a non-nil object for %q", doi)
		}
		record, err := CrosswalkDataCiteObject(app.Cfg, obj, options)
		if err != nil {
			t.Errorf("%s", err)
			t.FailNow()
		}
		if record == nil {
			t.Errorf("expected a non-nil record for doi %q", doi)
			t.FailNow()
		}
		/*
		optionsFName, diffFName := "", ""
		if err := app.RunDataCiteToRdm(os.Stdin, os.Stdout, os.Stderr, optionsFName, doi, diffFName); err != nil {
			t.Error(err)
			t.FailNow()
		}
		*/
	}
}
