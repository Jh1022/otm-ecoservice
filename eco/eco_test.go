package eco

import (
	"math/rand"
	"testing"
)

// Data sanity checks
func TestRegionsArePresent(t *testing.T) {
	// Spot check NoEastXXX and LoMidWXXX
	m, _ := LoadSpeciesMap("../data/species.json")

	_, contained := m["NoEastXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}

	_, contained = m["LoMidWXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}

	// Also check in base dictionary
	l := LoadFiles("../data/")

	_, contained = l["NoEastXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}

	_, contained = l["LoMidWXXX"]

	if !contained {
		t.Fatal("Missing NoEastXXX in master species table")
	}
}

func TestDataMappings(t *testing.T) {
	otmcode := "MASO"

	expected := map[string]string{
		"PiedmtCLT": "BDS OTHER",
		"NoEastXXX": "BDS OTHER",
		"CaNCCoJBK": "BDS OTHER",
		"InlValMOD": "MAGR",
		"SoCalCSMA": "BDS OTHER",
		"GulfCoCHS": "BDS OTHER",
		"CenFlaXXX": "BDS OTHER",
		"PacfNWLOG": "BDS OTHER",
		"InlEmpCLM": "MAGR",
	}

	m, _ := LoadSpeciesMap("../data/species.json")

	for region, target := range expected {
		regionmap, found := m[region]

		if !found {
			t.Fatalf("Missing region %v", region)
		}

		itreecode, found := regionmap[otmcode]

		if !found {
			t.Fatalf("Missing data for code %v in region %v",
				itreecode, region)
		}

		if itreecode != target {
			t.Fatalf("Invalid itree code for otmcode %v "+
				"in region %v (got %v, expected %v)",
				itreecode, region, itreecode, target)
		}
	}
}

func TestSimpleInter(t *testing.T) {
	breaks := []float64{1.0, 3.0}
	values := []float64{4.0, 6.0}

	itreecode := "blah"

	datafile := &Datafile{breaks, map[string][]float64{itreecode: values}}
	datafiles := []*Datafile{datafile}

	result := []float64{0.0}

	CalcOneTree(
		datafiles,
		itreecode,
		2.0,
		result)

	if result[0] != 5.0 {
		t.Fatalf("Expected %v, got %v", 2.0, result[0])
	}
}

// Since there isn't really a canonical benefits
// library to test against, we're just going to
// use a couple of exsiting OTM instances
//
// These test are pretty data dependent but the
// whole purpose of the library is to provide access to
// that data
func TestSpecificTreeData(t *testing.T) {
	region := "LoMidWXXX"
	otmcode := "ULAM"
	dbh := 11.0

	targets := map[string]float64{
		"aq_nox_avoided":     0.01548490,
		"aq_nox_dep":         0.00771784,
		"aq_pm10_avoided":    0.00546863,
		"aq_pm10_dep":        0.016322,
		"aq_sox_avoided":     0.06590,
		"aq_sox_dep":         0.0057742,
		"aq_voc_avoided":     0.0054686,
		"bvoc":               0,
		"co2_avoided":        12.0864829,
		"co2_sequestered":    51.42926,
		"co2_storage":        110.79107,
		"electricity":        12.180839,
		"hydro_interception": 2.5919028,
		"natural_gas":        -18.345013}

	l := LoadFiles("../data/")
	m, _ := LoadSpeciesMap("../data/species.json")

	itreecode := m[region][otmcode]
	factorDataForRegion := l[region]

	factorsum := make([]float64, len(Factors))

	CalcOneTree(
		factorDataForRegion,
		itreecode,
		dbh,
		factorsum)

	factormap := FactorArrayToMap(factorsum)

	for factor, target := range targets {
		calcd := factormap[factor]

		if int(calcd*100000) != int(target*100000) {
			t.Fatalf("Expected %v, got %v for factor %v",
				target, calcd, factor)
		}
	}
}

func generateSpeciesListFromRegion(
	speciesdata map[string]map[string]string,
	targetLength int,
	region string) []*TestRecord {

	speciesmap := speciesdata[region]

	possibleSpecies := make([]string, len(speciesmap))
	i := 0
	for k, _ := range speciesmap {
		possibleSpecies[i] = k
		i++
	}

	idx := rand.Perm(targetLength)

	data := make([]*TestRecord, targetLength)

	i = 0
	for sidx := range idx {
		otmcode := possibleSpecies[sidx%len(possibleSpecies)]
		diameter := rand.Float64() * 100.0
		data[i] = &TestRecord{otmcode, diameter, region}
		i++
	}

	return data
}

func benchmarkTreesSingleRegion(targetLength int, b *testing.B) {
	region := "LoMidWXXX"

	benchmarkTreesMultiRegion([]string{region}, targetLength, b)
}

func benchmarkTreesMultiRegion(regions []string, targetLength int, b *testing.B) {
	l := LoadFiles("../data/")
	speciesdata, _ := LoadSpeciesMap("../data/species.json")

	targetLengthPerRegion := targetLength / len(regions)
	data := make([]*TestRecord, 0)

	for i := range regions {
		newdata := generateSpeciesListFromRegion(
			speciesdata, targetLengthPerRegion, regions[i])
		data = append(data, newdata...)
	}

	testingContext := &TestingContext{
		len(regions) > 1, regions[0], 0, data}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testingContext.Reset()
		data, err := CalcBenefits(testingContext, 0, speciesdata, l)

		if err != nil {
			b.Fatalf("error: %v", err)
		}
		benchdump = data
	}

}

// Store stuff to this variable to prevent the compiler
// from getting to tricky
var (
	benchdump interface{}
	regions   []string = []string{"PiedmtCLT", "NoEastXXX", "CaNCCoJBK",
		"InlValMOD", "SoCalCSMA", "GulfCoCHS",
		"CenFlaXXX", "PacfNWLOG", "InlEmpCLM"}
)

func BenchmarkTreesMultiRegion100(b *testing.B) {
	benchmarkTreesMultiRegion(regions, 1e2, b)
}

func BenchmarkTreesMultiRegion1k(b *testing.B) {
	benchmarkTreesMultiRegion(regions, 1e3, b)
}

func BenchmarkTreesMultiRegion10k(b *testing.B) {
	benchmarkTreesMultiRegion(regions, 1e4, b)
}

func BenchmarkTreesMultiRegion100k(b *testing.B) {
	benchmarkTreesMultiRegion(regions, 1e5, b)
}

func BenchmarkTreesMultiRegion1M(b *testing.B) {
	benchmarkTreesMultiRegion(regions, 1e6, b)
}

func BenchmarkTreesSingleRegion100(b *testing.B)  { benchmarkTreesSingleRegion(1e2, b) }
func BenchmarkTreesSingleRegion1k(b *testing.B)   { benchmarkTreesSingleRegion(1e3, b) }
func BenchmarkTreesSingleRegion10k(b *testing.B)  { benchmarkTreesSingleRegion(1e4, b) }
func BenchmarkTreesSingleRegion100k(b *testing.B) { benchmarkTreesSingleRegion(1e5, b) }
func BenchmarkTreesSingleRegion1M(b *testing.B)   { benchmarkTreesSingleRegion(1e6, b) }
