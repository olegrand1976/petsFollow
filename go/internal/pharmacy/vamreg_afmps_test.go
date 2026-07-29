package pharmacy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

func TestVamregAFMPSClientRequiresAPIKey(t *testing.T) {
	c := pharmacy.NewVamregAFMPSClient("http://example.invalid", "")
	_, err := c.ListUnits(context.Background())
	if err == nil || !strings.Contains(err.Error(), "vamreg_afmps_api_key_required") {
		t.Fatalf("want api key error, got %v", err)
	}
}

func TestVamregAFMPSClientMedicinalProductsAndHeader(t *testing.T) {
	var gotPath, gotKey, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get(pharmacy.VamregAFMPSSecKeyHeader)
		gotAccept = r.Header.Get("Accept")
		if r.URL.Query().Get("timestamp") == "true" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"timestamp":"2026-01-21T13:15:16.360322483Z",
				"medicinalProducts":[{
					"cti":"232346-02","cnk":"1691716",
					"ppnEn":"Advocin","ppnNl":"Advocin","ppnFr":"Advocin",
					"packSize":"100 ml","usage":"VETERINARY",
					"activeSubstanceStrengthUnits":[{"activeSubstanceSporId":"100000083438","strengthUnitLabel":"180 mg/ml"}],
					"pharmaceuticalFormSporIds":["100000073863"],
					"maName":"Advocin 180","maNumber":"BE-V232346","mahName":"Zoetis Belgium",
					"deprecated":false,"amrDeclarationUnit":"G","amrPackUnitAmount":"100"
				}]
			}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"cti":"1","cnk":"2","ppnEn":"x","ppnNl":"x","ppnFr":"x","packSize":"1","usage":"VETERINARY","amrPackUnitAmount":1.5,"amrDeclarationUnit":"ML","deprecated":false}]`))
	}))
	defer srv.Close()

	c := pharmacy.NewVamregAFMPSClient(srv.URL, "key-uuid_secret=")
	c.HTTPClient = srv.Client()

	list, err := c.ListMedicinalProducts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != pharmacy.VamregPathMedicinalProduct || gotKey != "key-uuid_secret=" || gotAccept != "application/json" {
		t.Fatalf("path=%q key=%q accept=%q", gotPath, gotKey, gotAccept)
	}
	if len(list) != 1 || list[0].CNK != "2" || list[0].AmrPackUnitAmount.Float64() != 1.5 {
		t.Fatalf("list=%#v", list)
	}

	timed, err := c.ListMedicinalProductsWithTimestamp(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(timed.MedicinalProducts) != 1 || timed.MedicinalProducts[0].CNK != "1691716" {
		t.Fatalf("timed=%#v", timed)
	}
	if timed.MedicinalProducts[0].AmrPackUnitAmount.Float64() != 100 {
		t.Fatalf("string amount=%v", timed.MedicinalProducts[0].AmrPackUnitAmount)
	}
}

func TestVamregAFMPSClientForeignTimedFieldVariants(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"timestamp":"2026-01-21T13:15:16Z",
			"nonBeMedicinalProducts":[{
				"foreignProductName":"SPIROVET","foreignMaNumber":"FR/1","foreignMaHolder":"Ceva",
				"pharmaceuticalFormSporId":"100000073863","packSize":"100 ml","usage":"VETERINARY",
				"updPermanentIdentifier":"600000040277","updPackId":"129540fc-1a0f-439a-a45c-2f65c2b38218",
				"deprecated":false,
				"activeSubstanceStrengthUnitMasses":[{"activeSubstanceSporId":"100000091325","strength":"6E+5","unitSporId":"100000110736"}]
			}]
		}`))
	}))
	defer srv.Close()

	c := pharmacy.NewVamregAFMPSClient(srv.URL, "k")
	c.HTTPClient = srv.Client()
	timed, err := c.ListForeignMedicinalProductsWithTimestamp(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	prods := timed.Products()
	if len(prods) != 1 || prods[0].ForeignProductName != "SPIROVET" {
		t.Fatalf("%#v", timed)
	}
}

func TestVamregAFMPSClientForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	c := pharmacy.NewVamregAFMPSClient(srv.URL, "bad")
	c.HTTPClient = srv.Client()
	_, err := c.ListActiveSubstances(context.Background())
	if err == nil || !strings.Contains(err.Error(), "vamreg_afmps_forbidden") {
		t.Fatalf("got %v", err)
	}
}

func TestVamregAFMPSClientCodeLists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"code":"DOG","en":"Dog","nl":"Hond","fr":"Chien","deprecated":false}]`))
	}))
	defer srv.Close()
	c := pharmacy.NewVamregAFMPSClient(srv.URL, "k")
	c.HTTPClient = srv.Client()
	for _, fn := range []func(context.Context) ([]pharmacy.VamregCodeLabel, error){
		c.ListActiveSubstances,
		c.ListPharmaceuticalForms,
		c.ListUnits,
		c.ListUnitsOut,
		c.ListTargetSpecies,
		c.ListIndications,
		c.ListProviderTypes,
		c.ListProductTypes,
		c.ListUsages,
	} {
		rows, err := fn(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || rows[0].Code != "DOG" {
			t.Fatalf("%#v", rows)
		}
	}
}

func TestNewVamregAFMPSClientDefaultBase(t *testing.T) {
	c := pharmacy.NewVamregAFMPSClient("", "k")
	if c.BaseURL != pharmacy.VamregAFMPSDefaultBaseURL {
		t.Fatalf("base=%q", c.BaseURL)
	}
}
