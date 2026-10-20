package organize

import (
	"path/filepath"
	"sort"
	"strings"
)

// GraphClusterer groups files based on mutual imports and topological layers.
type GraphClusterer struct {
	Files map[string]*FileNode
}

// NewGraphClusterer initializes a new clusterer instance.
func NewGraphClusterer(nodes []*FileNode) *GraphClusterer {
	filesMap := make(map[string]*FileNode)
	for _, n := range nodes {
		filesMap[n.RelPath] = n
	}
	return &GraphClusterer{Files: filesMap}
}

// ClusterByLabelPropagation runs deterministic Label Propagation on code files.
func (gc *GraphClusterer) ClusterByLabelPropagation(maxIterations int) map[string]string {
	labels := make(map[string]string)
	var codePaths []string

	for path, node := range gc.Files {
		if node.Category == CategoryCode {
			// Base label is the file stem or current package name
			base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			if node.PackageName != "" && node.PackageName != "main" {
				base = node.PackageName
			}
			labels[path] = base
			codePaths = append(codePaths, path)
		}
	}

	sort.Strings(codePaths)

	// Build bidirectional adjacency weights
	weights := make(map[string]map[string]int)
	for _, u := range codePaths {
		if weights[u] == nil {
			weights[u] = make(map[string]int)
		}
		for _, imp := range gc.Files[u].Imports {
			// Check if imported target matches another workspace file
			for _, v := range codePaths {
				if strings.Contains(imp, strings.TrimSuffix(filepath.Base(v), filepath.Ext(v))) {
					weights[u][v] += 2
					if weights[v] == nil {
						weights[v] = make(map[string]int)
					}
					weights[v][u] += 2
				}
			}
		}
	}

	// Iterative propagation
	for iter := 0; iter < maxIterations; iter++ {
		changed := false

		for _, u := range codePaths {
			neighborLabelWeights := make(map[string]int)
			for v, w := range weights[u] {
				currLabel := labels[v]
				neighborLabelWeights[currLabel] += w
			}

			bestLabel := labels[u]
			maxWeight := 0

			var candidateLabels []string
			for l := range neighborLabelWeights {
				candidateLabels = append(candidateLabels, l)
			}
			sort.Strings(candidateLabels)

			for _, l := range candidateLabels {
				w := neighborLabelWeights[l]
				if w > maxWeight {
					maxWeight = w
					bestLabel = l
				}
			}

			if bestLabel != labels[u] && maxWeight > 0 {
				labels[u] = bestLabel
				changed = true
			}
		}

		if !changed {
			break
		}
	}

	return labels
}

// TarjanDetector executes Tarjan's Strongly Connected Components algorithm.
type TarjanDetector struct {
	index   int
	stack   []string
	indices map[string]int
	lowlink map[string]int
	onStack map[string]bool
	edges   map[string][]string
	sccs    [][]string
}

// DetectCycles finds all cyclic dependencies in a package or file graph.
func DetectCycles(nodes []string, edges map[string][]string) [][]string {
	td := &TarjanDetector{
		indices: make(map[string]int),
		lowlink: make(map[string]int),
		onStack: make(map[string]bool),
		edges:   edges,
	}

	sort.Strings(nodes)
	for _, node := range nodes {
		if _, visited := td.indices[node]; !visited {
			td.strongConnect(node)
		}
	}

	var cycles [][]string
	for _, scc := range td.sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
		}
	}
	return cycles
}

func (td *TarjanDetector) strongConnect(u string) {
	td.indices[u] = td.index
	td.lowlink[u] = td.index
	td.index++
	td.stack = append(td.stack, u)
	td.onStack[u] = true

	for _, v := range td.edges[u] {
		if _, visited := td.indices[v]; !visited {
			td.strongConnect(v)
			if td.lowlink[v] < td.lowlink[u] {
				td.lowlink[u] = td.lowlink[v]
			}
		} else if td.onStack[v] {
			if td.indices[v] < td.lowlink[u] {
				td.lowlink[u] = td.indices[v]
			}
		}
	}

	if td.lowlink[u] == td.indices[u] {
		var scc []string
		for {
			w := td.stack[len(td.stack)-1]
			td.stack = td.stack[:len(td.stack)-1]
			td.onStack[w] = false
			scc = append(scc, w)
			if w == u {
				break
			}
		}
		td.sccs = append(td.sccs, scc)
	}
}

// GenerateTargetDest determines the canonical target directory for a file node based on preset.
func GenerateTargetDest(node *FileNode, preset string, clusterName string) (string, string) {
	base := filepath.Base(node.RelPath)

	// 1. Root Anchors stay at root
	if node.Category == CategoryRootAnchor {
		return node.RelPath, ""
	}

	// 2. Non-code assets
	switch node.Category {
	case CategoryDocs:
		return filepath.ToSlash(filepath.Join("docs", base)), ""
	case CategoryAssets:
		return filepath.ToSlash(filepath.Join("assets", base)), ""
	case CategoryScripts:
		return filepath.ToSlash(filepath.Join("scripts", base)), ""
	case CategoryConfigs:
		return filepath.ToSlash(filepath.Join("configs", base)), ""
	case CategoryTestData:
		return filepath.ToSlash(filepath.Join("testdata", base)), ""
	}

	// 3. Source Code files
	switch preset {
	case "clean-arch":
		switch node.Role {
		case RoleEntry:
			return filepath.ToSlash(filepath.Join("infra", "cmd", base)), "main"
		case RoleAPI:
			return filepath.ToSlash(filepath.Join("adapters", "api", base)), "api"
		case RoleModel:
			return filepath.ToSlash(filepath.Join("domain", base)), "domain"
		case RoleService:
			return filepath.ToSlash(filepath.Join("usecases", base)), "usecases"
		case RoleUtil:
			return filepath.ToSlash(filepath.Join("infra", "utils", base)), "utils"
		default:
			return filepath.ToSlash(filepath.Join("domain", base)), "domain"
		}

	case "standard-go", "default", "":
		fallthrough
	default:
		switch node.Role {
		case RoleEntry:
			stem := strings.TrimSuffix(base, filepath.Ext(base))
			if stem == "main" {
				stem = "app"
			}
			return filepath.ToSlash(filepath.Join("cmd", stem, base)), "main"
		case RoleAPI:
			return filepath.ToSlash(filepath.Join("internal", "api", base)), "api"
		case RoleModel:
			return filepath.ToSlash(filepath.Join("internal", "models", base)), "models"
		case RoleService:
			pkg := "services"
			if clusterName != "" && clusterName != base {
				pkg = clusterName
			}
			return filepath.ToSlash(filepath.Join("internal", pkg, base)), pkg
		case RoleUtil:
			return filepath.ToSlash(filepath.Join("pkg", "utils", base)), "utils"
		default:
			if clusterName != "" {
				return filepath.ToSlash(filepath.Join("internal", clusterName, base)), clusterName
			}
			return filepath.ToSlash(filepath.Join("internal", "core", base)), "core"
		}
	}
}
