package proxy

func ReverseString(req string) string {
	if req[len(req)-1:] != "." {
		req = req + "."
	}
	rns := []rune(req)
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		rns[i], rns[j] = rns[j], rns[i]
	}
	return string(rns)
}
func (p *Proxy) updateRadixTree() {
	for k, v := range p.route {
		p.routeTree.Insert(ReverseString(k), v)
	}

}

func (p *Proxy) DomainRouteLookup(req string) []string {
	result := make([]string, 0)
	_, v, found := p.routeTree.LongestPrefix(ReverseString(req))
	if found {
		return v.([]string)
	}
	return result
}
