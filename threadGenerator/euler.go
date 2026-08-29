package threadGenerator

// eulerTour returns a consecutive nail walk that covers every undirected edge
// at least once (Hierholzer). Odd-degree vertices are paired with extra edges
// so the graph is Eulerian, matching StringArt's findConsecutivePath extras.
func eulerTour(n, start int, edges [][2]int) []Path {
	if n <= 0 || len(edges) == 0 {
		return nil
	}
	adj := make([][]int, n)
	add := func(a, b int) {
		adj[a] = append(adj[a], b)
		adj[b] = append(adj[b], a)
	}
	for _, e := range edges {
		if e[0] >= 0 && e[1] >= 0 && e[0] < n && e[1] < n && e[0] != e[1] {
			add(e[0], e[1])
		}
	}
	odd := make([]int, 0, 8)
	for i := 0; i < n; i++ {
		if len(adj[i])%2 == 1 {
			odd = append(odd, i)
		}
	}
	for i := 0; i+1 < len(odd); i += 2 {
		add(odd[i], odd[i+1])
	}

	if start < 0 || start >= n || len(adj[start]) == 0 {
		start = 0
		for i := 0; i < n; i++ {
			if len(adj[i]) > 0 {
				start = i
				break
			}
		}
	}

	removeOne := func(from, to int) {
		for i, v := range adj[from] {
			if v == to {
				adj[from] = append(adj[from][:i], adj[from][i+1:]...)
				return
			}
		}
	}

	stack := []int{start}
	circuit := make([]int, 0, len(edges)+n+1)
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		if len(adj[u]) > 0 {
			v := adj[u][len(adj[u])-1]
			adj[u] = adj[u][:len(adj[u])-1]
			removeOne(v, u)
			stack = append(stack, v)
			continue
		}
		circuit = append(circuit, u)
		stack = stack[:len(stack)-1]
	}
	for i, j := 0, len(circuit)-1; i < j; i, j = i+1, j-1 {
		circuit[i], circuit[j] = circuit[j], circuit[i]
	}
	if len(circuit) < 2 {
		return nil
	}
	out := make([]Path, 0, len(circuit)-1)
	for i := 0; i+1 < len(circuit); i++ {
		if circuit[i] == circuit[i+1] {
			continue
		}
		out = append(out, Path{StartingNail: circuit[i], EndingNail: circuit[i+1]})
	}
	return out
}
