package git

import (
	"testing"
)

func TestReachableFromAny(t *testing.T) {
	repo, err := OpenRepository("testdata/TestGitRepository.git")
	checkFatal(t, err)
	defer repo.Free()

	for name, tc := range map[string]struct {
		reachable   bool
		commit      string
		descendants []string
	}{
		"empty": {
			reachable: false,
			commit:    "49322bb17d3acc9146f98c97d078513228bbf3c0",
		},
		"same": {
			reachable:   true,
			commit:      "49322bb17d3acc9146f98c97d078513228bbf3c0",
			descendants: []string{"49322bb17d3acc9146f98c97d078513228bbf3c0"},
		},
		"unreachable": {
			reachable:   false,
			commit:      "ac7e7e44c1885efb472ad54a78327d66bfc4ecef",
			descendants: []string{"58be4659bb571194ed4562d04b359d26216f526e"},
		},
		"unreachable-reverse": {
			reachable:   false,
			commit:      "58be4659bb571194ed4562d04b359d26216f526e",
			descendants: []string{"ac7e7e44c1885efb472ad54a78327d66bfc4ecef"},
		},
		"root": {
			reachable: false,
			commit:    "42e4e7c5e507e113ebbb7801b16b52cf867b7ce1",
			descendants: []string{
				"ac7e7e44c1885efb472ad54a78327d66bfc4ecef",
				"d86a2aada2f5e7ccf6f11880bfb9ab404e8a8864",
				"f73b95671f326616d66b2afb3bdfcdbbce110b44",
				"d0114ab8ac326bab30e3a657a0397578c5a1af88",
			},
		},
		"head": {
			reachable: false,
			commit:    "49322bb17d3acc9146f98c97d078513228bbf3c0",
			descendants: []string{
				"ac7e7e44c1885efb472ad54a78327d66bfc4ecef",
				"d86a2aada2f5e7ccf6f11880bfb9ab404e8a8864",
				"f73b95671f326616d66b2afb3bdfcdbbce110b44",
				"d0114ab8ac326bab30e3a657a0397578c5a1af88",
			},
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			commit, err := NewOid(tc.commit)
			checkFatal(t, err)

			descendants := make([]*Oid, len(tc.descendants))
			for i, o := range tc.descendants {
				descendants[i], err = NewOid(o)
				checkFatal(t, err)
			}
			reachable, err := repo.ReachableFromAny(commit, descendants)
			checkFatal(t, err)

			if reachable != tc.reachable {
				t.Errorf("ReachableFromAny(%s, %v) = %v, wanted %v", tc.commit, tc.descendants, reachable, tc.reachable)
			}
		})
	}
}
