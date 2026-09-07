package main

// oracle_gen_test emits the oracle passage map for Step 3's Qwen/oracle cell: for each of the 5 real
// claims, the passages the Sonnet N=3 defender actually quoted (looked up in the real corpus, not
// transcribed), plus the three mutants mapped to their base claim's passages (M1←F31, M2←F17,
// M3←F12). Gated by EMIT_ORACLE=1 so it never runs in the normal suite; the JSON it writes is
// committed and read by `assay -retrieve=oracle`.

import (
	"encoding/json"
	"os"
	"testing"

	"assay/internal/retrieve"
)

func TestEmitOracle(t *testing.T) {
	if os.Getenv("EMIT_ORACLE") != "1" {
		t.Skip("set EMIT_ORACLE=1 to regenerate examples/vic-lceic/oracle-18-spans.json")
	}
	ix, err := retrieve.Load("examples/vic-lceic/sources/hearings")
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	claims := readClaimLines(t, "examples/vic-lceic/claims-faith-5.txt") // F8 F12 F17 F29 F31
	quotes := readChainQuotes(t, "eval/n3/claims-faith-refuter.faithfulness.jsonl")

	oracle := map[string][]string{}
	for i, cl := range claims {
		seen := map[string]bool{}
		var ids []string
		for _, q := range quotes[i] {
			if pid := passageContaining(ix, q); pid != "" && !seen[pid] {
				seen[pid] = true
				ids = append(ids, pid)
			}
		}
		if len(ids) == 0 {
			t.Fatalf("claim %s: no gold passages located", cl.id)
		}
		oracle[cl.id] = ids
	}
	// Mutants share their base claim's gold passages: the perfect evidence still refutes the mutant.
	for mutant, base := range map[string]string{"M1": "F31", "M2": "F17", "M3": "F12"} {
		oracle[mutant] = oracle[base]
	}

	data, _ := json.MarshalIndent(oracle, "", "  ")
	if err := os.WriteFile("examples/vic-lceic/oracle-18-spans.json", data, 0o644); err != nil {
		t.Fatalf("write oracle: %v", err)
	}
	t.Logf("wrote oracle for %d claims", len(oracle))
}
