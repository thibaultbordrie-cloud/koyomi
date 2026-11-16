// Package guard holds the anonymity test: nothing in this public repository may point to a
// real company, person, internal tool or cloud project. The forbidden terms are stored only
// as hashes, so this file does not reveal them either. E-mail addresses must be example ones.
package guard

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var forbidden = map[string]bool{
	"08f6f59e0540944622c9": true, "0a13fe27742142ab0a55": true, "2384ec0e2963ba2a2c71": true, "3462ee2ec1e78c77d7f0": true,
	"3af7d17d9a4bc5308c57": true, "43b54885c5690ed70656": true, "44b1d90c896f767137f4": true, "54557012f65abc5908f3": true,
	"5845901d0e075c9bb3ab": true, "69185407befe30912e51": true, "7cbdf191488e3c89e8ce": true, "844ebbe693aaaf6a39d3": true,
	"915f4a517d06ef5fbf31": true, "91ae1deae4b736c50156": true, "9768f993f262761fe41e": true, "9be83f38bdcd57c6fe98": true,
	"abb81aaf9978b9966fed": true,
}

var (
	accents = strings.NewReplacer("à", "a", "â", "a", "ä", "a", "ç", "c", "é", "e", "è", "e", "ê", "e", "ë", "e",
		"î", "i", "ï", "i", "ô", "o", "ö", "o", "ù", "u", "û", "u", "ü", "u")
	word = regexp.MustCompile(`[a-z0-9]+`)
	mail = regexp.MustCompile(`[\w.+-]+@([\w-]+(?:\.[\w-]+)+)`)
)

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:20]
}

func files(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == "site" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(path)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 15 {
		t.Fatalf("only %d files scanned", len(out))
	}
	return out
}

func TestNoIdentifyingTerm(t *testing.T) {
	for path, text := range files(t) {
		words := word.FindAllString(accents.Replace(strings.ToLower(text+" "+path)), -1)
		for i := range words {
			for n := 1; n <= 3 && i+n <= len(words); n++ {
				if forbidden[hash(strings.Join(words[i:i+n], " "))] {
					t.Errorf("identifying term found in %s", path)
				}
			}
		}
	}
}

func TestOnlyExampleEmailAddresses(t *testing.T) {
	for path, text := range files(t) {
		for _, m := range mail.FindAllStringSubmatch(text, -1) {
			domain := strings.ToLower(m[1])
			if !strings.HasSuffix(domain, "example.com") && !strings.HasSuffix(domain, "example.org") &&
				!strings.HasSuffix(domain, "users.noreply.github.com") {
				t.Errorf("%s: e-mail address at %s", path, domain)
			}
		}
	}
}
