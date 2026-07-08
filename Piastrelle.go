package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	p := newPiano()
	inputSys(p)
}

//tipi

type piano struct {
	tiles map[coordinate]*tile
	rules *list[rule]
}

type coordinate struct {
	x int
	y int
}

type rule struct {
	risultato string
	operands  *list[propOperand]
	consumo   int
}

type propOperand struct {
	k int
	a string
}

type tile struct {
	color     string
	intensity int
	adj       *list[coordinate]
}

type node[T comparable] struct {
	item T
	next *node[T]
	prev *node[T]
}

type list[T comparable] struct {
	head *node[T]
}

//funzioni e metodi

// esegue per il piano p l'operazione specificata nella stringa s
func esegui(p piano, s string) {
	var spl []string //contiene i singoli argomenti in forma stringa, la maggior parte delle funzioni la usa
	var x, y int     //contiene gli argomenti x e y, vengono usati in tutte le funzioni che richiedono split quindi vale la pena gestire qui anche queste
	if s[0] != 'r' && s[0] != 's' && s[0] != 'o' && s[0] != 'q' {
		spl = strings.Split(s[2:], " ")
		if len(spl) >= 2 {
			x, _ = strconv.Atoi(spl[0])
			y, _ = strconv.Atoi(spl[1])
		}
	}
	switch s[0] {
	case 'C':
		i, _ := strconv.Atoi(spl[3])
		colora(p, x, y, spl[2], i)
	case 'S':
		spegni(p, x, y)
	case 'r':
		regola(p, s[2:])
	case '?':
		stato(p, x, y)
	case 's':
		stampa(p)
	case 'b':
		blocco(p, x, y)
	case 'B':
		bloccoOmog(p, x, y)
	case 'p':
		propaga(p, x, y)
	case 'P':
		propagaBlocco(p, x, y)
	case 'o':
		ordina(p)
	case 'q':
		os.Exit(0)
	default:

	}
}

// colora la piastrella(x,y) del colore alpha qualunque fosse il suo stato (se è spenta si accende, se è accesa cambia colore)
func colora(p piano, x int, y int, alpha string, i int) {
	if i == 0 {
		spegni(p, x, y)
		return
	}
	if p.getCol(x, y) != "" {
		p.tiles[coord(x, y)] = &tile{color: alpha, intensity: i, adj: p.getAdj(x, y)}
	} else {
		p.tiles[coord(x, y)] = &tile{color: alpha, intensity: i, adj: &list[coordinate]{}}
		for i := -1; i < 2; i++ {
			for j := -1; j < 2; j++ {
				if i == 0 && j == 0 {
					continue
				}
				if p.getCol(x+j, y+i) != "" {
					addNode(p.getAdj(x, y), coord(x+j, y+i))
					addNode(p.getAdj(x+j, y+i), coord(x, y))
				}
			}
		}
	}
}

// spegne la piastrella(x,y), se è già spenta non fa nulla
func spegni(p piano, x int, y int) {
	pos := coord(x, y)
	if p.getCol(x, y) != "" {
		if p.getAdj(x, y).head != nil {
			curr := p.getAdj(x, y).head
			delNode(p.getAdj(curr.item.x, curr.item.y), pos)
			curr = curr.next
			for curr != p.getAdj(x, y).head {
				delNode(p.getAdj(curr.item.x, curr.item.y), pos)
				curr = curr.next
			}
		}
		delete(p.tiles, pos)
	}
}

// Definisce la regola di propagazione k1α1 + k2α2 + · · · + knαn → β e la inserisce in fondo all’elenco delle regole
func regola(p piano, r string) {
	in := strings.Split(r, " ")
	nRule := rule{risultato: in[0], consumo: 0, operands: &list[propOperand]{}}
	for i := 1; i < len(in); i += 2 {
		k, _ := strconv.Atoi(in[i])
		addNode(nRule.operands, propOperand{k: k, a: in[i+1]})
	}
	addNode(p.rules, nRule)
}

// Stampa e restituisce il colore e l’intensità di Piastrella(x, y). Se Piastrella(x, y) è spenta, non stampa nulla
func stato(p piano, x int, y int) (string, int) {
	col := p.getCol(x, y)
	inten := p.getInten(x, y)
	if col != "" {
		fmt.Println("colore:", col, "intensità:", inten)
	}
	return col, inten
}

// Stampa l’elenco delle regole di propagazione, nell’ordine attuale
func stampa(p piano) {
	println("(")
	if p.rules != nil && p.rules.head != nil {
		curr := p.rules.head
		printRule(curr.item)
		curr = curr.next
		for curr != p.rules.head {
			printRule(curr.item)
			curr = curr.next
		}
	}
	println(")")
}

// Calcola e stampa la somma delle intensità delle piastrelle contenute nel blocco di appartenenza di Piastrella(x, y). Se Piastrella(x, y) è spenta, restituisce 0
func blocco(p piano, x int, y int) {
	b := make(map[coordinate]*tile)
	_, i := findBlock(p, x, y, "", b, 0)
	fmt.Println(i)
}

// Calcola e stampa la somma delle intensità delle piastrelle contenute nel blocco omogeneo di appartenenza di Piastrella(x, y). Se Piastrella(x, y) è spenta, restituisce 0
func bloccoOmog(p piano, x int, y int) {
	_, i := findBlock(p, x, y, p.getCol(x, y), make(map[coordinate]*tile), 0)
	fmt.Println(i)
}

// Applica a Piastrella(x, y) la prima regola di propagazione applicabile dell’elenco, ricolorando la piastrella. Se nessuna regola è applicabile, non viene eseguita alcuna operazione
func propaga(p piano, x int, y int) {
	var state bool //per vedere se posso applicare una regola la piastrella deve avere una lista di adiacenza, quindi deve essere accesa, questa variabile tiene traccia dello stato precedente, se nessuna regola è applicabile e stato è false spengo la piastrella al termine dell'esecuzione di propaga
	if p.getCol(x, y) == "" {
		state = false
		colora(p, x, y, "0", 1)
	} else {
		state = true
	}
	res := findValidRule(p, x, y)
	if res != "" {
		colora(p, x, y, res, p.getInten(x, y))
		return
	}
	if !state {
		spegni(p, x, y)
	}
}

// Applica su ogni piastrella sul blocco di appartenenza di Piastrella(x, y) la prima regola applicabile a ognuna
func propagaBlocco(p piano, x int, y int) {
	b, _ := findBlock(p, x, y, "", make(map[coordinate]*tile), 0)
	risultati := make(map[coordinate]string)
	for i := range b {
		res := findValidRule(p, i.x, i.y)
		if res != "" {
			risultati[i] = res
		}
	}
	for i, v := range risultati {
		b[i].color = v
	}
}

// Ordina l’elenco delle regole di propagazione in base al consumo delle regole stesse: la regola con consumo maggiore diventa l’ultima dell’elenco.
// Se due regole hanno consumo uguale mantengono il loro ordine relativo, utilizza l'algoritmo insertion sort
func ordina(p piano) {
	sl := listToSlice(p.rules)
	for i := 1; i < len(sl); i++ {
		x := sl[i]
		j := i - 1
		for j >= 0 && sl[j].consumo > x.consumo {
			sl[j], sl[j+1] = sl[j+1], sl[j]
			j--
		}
		sl[j+1] = x
	}
	p.rules.head = sliceToList(sl).head
	return
}

// sistema di input, prende una riga in input da standard input ed esegue il comando associato
func inputSys(p piano) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		/*if s.Text()[0] == 'q' {
			return
		}*/
		esegui(p, s.Text())
	}
}

// crea un piano con nessuna piastrella accesa e nessuna regola di propagazione
func newPiano() piano {
	return piano{tiles: make(map[coordinate]*tile), rules: &list[rule]{}}
}

// converte due interi in tipo coordinate, permette una sintassi semplificata per le coordinate
func coord(x int, y int) coordinate {
	return coordinate{x: x, y: y}
}

// restituisce il puntatore della piastrella alle coordinate x,y del piano p
func (p piano) getTile(x, y int) *tile {
	return p.tiles[coord(x, y)]
}

// restituisce il colore della piastrella alle coordinate x,y del piano p
func (p piano) getCol(x, y int) string {
	if p.tiles[coord(x, y)] != nil {
		return p.tiles[coord(x, y)].color
	}
	return ""
}

// restituisce l'intensità della piastrella alle coordinate x,y del piano p
func (p piano) getInten(x, y int) int {
	if p.tiles[coord(x, y)] != nil {
		return p.tiles[coord(x, y)].intensity
	}
	return 0
}

// restituisce il puntatore alla lista di adiacenza della piastrella alle coordinate x,y del piano p
func (p piano) getAdj(x, y int) *list[coordinate] {
	return p.tiles[coord(x, y)].adj
}

// crea un nuovo nodo per le liste
func newNode[T comparable](val T) *node[T] {
	return &node[T]{val, nil, nil}
}

// aggiunge un nodo alla lista
func addNode[T comparable](l *list[T], val T) {
	node := newNode(val)
	if l.head == nil {
		l.head = node
		l.head.next = l.head
		l.head.prev = l.head
	} else {
		node.prev = l.head.prev
		node.next = l.head
		l.head.prev.next = node
		l.head.prev = node
	}
}

// rimouove un nodo dalla lista dato il suo valore
func delNode[T comparable](l *list[T], val T) {
	if l.head.item == val {
		if l.head == l.head.next {
			l = nil
			return
		} else {
			l.head = l.head.next
			l.head.prev = l.head.prev.prev
			l.head.prev.next = l.head
			return
		}
	}
	curr, prev := l.head, l.head.prev
	for curr.next != l.head {
		if curr.item == val {
			prev.next = curr.next
			curr.next.prev = prev
			return
		}
	}
}

// converte una lista in una slice equivalente
func listToSlice[T comparable](l *list[T]) []T {
	sl := make([]T, 0)
	if l == nil || l.head == nil {
		return sl
	}
	curr := l.head
	sl = append(sl, curr.item)
	curr = curr.next
	for curr != l.head {
		sl = append(sl, curr.item)
		curr = curr.next
	}
	return sl
}

// converte una slice in una lista equivalente
func sliceToList[T comparable](sl []T) *list[T] {
	l := list[T]{}
	for i := 0; i < len(sl); i++ {
		addNode(&l, sl[i])
	}
	return &l
}

// stampa una singola regola di propagazione
func printRule(r rule) {
	fmt.Print(r.risultato, ": ")
	if r.operands == nil || r.operands.head == nil {
		return
	}
	curr := r.operands.head
	fmt.Print(curr.item.k, " ")
	fmt.Print(curr.item.a, " ")
	curr = curr.next
	for curr != r.operands.head {
		fmt.Print(curr.item.k, " ")
		fmt.Print(curr.item.a, " ")
		curr = curr.next
	}
	println()
}

// data una regola di propagazione e una mappa con chiavi i colori e valori il numero di istanze di quel colore intorno a una piastrella
// controlla se la regola è applicabile
func checkRule(r rule, m map[string]int) bool {
	if r.operands != nil && r.operands.head != nil {
		curr := r.operands.head
		if m[curr.item.a] < curr.item.k {
			return false
		}
		curr = curr.next
		for curr != r.operands.head {
			if m[curr.item.a] < curr.item.k {
				return false
			}
			curr = curr.next
		}
	} else {
		return false
	}
	return true
}

// trova la prima regola applicabile per una data piastrella, ne aumenta il consumo di 1 e restituisce il colore del risultato della regola
// questa implementazione permette di utilizzare findValidRule sia per propaga che per propagaBlocco (nel primo caso si cerca una regola
// valida e se trovata in propaga si cambia il colore, nel secondo si esegue questa funzione per ogni piastrella del blocco e si salva
// ogni risultato con le rispettive coordinate in una struttura dati, dopo averla eseguita per ogni piastrella del blocco si applica
// ogni risultato
func findValidRule(p piano, x, y int) string {
	colCount := make(map[string]int)
	if p.getAdj(x, y) != nil && p.getAdj(x, y).head != nil {
		curr := p.getAdj(x, y).head
		colCount[p.getCol(curr.item.x, curr.item.y)]++
		curr = curr.next
		for curr != p.getAdj(x, y).head {
			colCount[p.getCol(curr.item.x, curr.item.y)]++
			curr = curr.next
		}
	}
	if p.rules != nil && p.rules.head != nil {
		curr := p.rules.head
		if checkRule(curr.item, colCount) {
			curr.item.consumo++
			return curr.item.risultato
		}
		curr = curr.next
		for curr != p.rules.head {
			if checkRule(curr.item, colCount) {
				curr.item.consumo++
				return curr.item.risultato
			}
			curr = curr.next
		}
	}
	return ""
}

// restituisce una mappa del blocco a cui appartiene la piastrella (x,y) e l'intensità dello stesso blocco, se c="" cerca un blocco normale, altrimenti cerca il blocco omogeneo di
// quel colore, se la piastrella è spenta o ha colore diverso da c (quando c!="") restituisce una mappa vuota, i contiene l'intensità del blocco (quando chiamata al di fuori delle
// chiamate ricorsive deve essere 0), utilizza un algoritmo DFS (Depth-First Search) ricorsivo
func findBlock(p piano, x, y int, c string, b map[coordinate]*tile, i int) (map[coordinate]*tile, int) {
	if i == 0 {
		if p.getCol(x, y) == "" {
			return b, i
		}
		b[coord(x, y)] = p.getTile(x, y)
		i = p.getInten(x, y)
	}
	if p.getAdj(x, y) != nil && p.getAdj(x, y).head != nil {
		curr := p.getAdj(x, y).head
		if b[curr.item] == nil && (c == "" || p.getCol(curr.item.x, curr.item.y) == c) {
			b[curr.item] = p.getTile(curr.item.x, curr.item.y)
			i += p.getInten(curr.item.x, curr.item.y)
			b, i = findBlock(p, curr.item.x, curr.item.y, c, b, i)
		}
		curr = curr.next
		for curr != p.getAdj(x, y).head {
			if b[curr.item] == nil && (c == "" || p.getCol(curr.item.x, curr.item.y) == c) {
				b[curr.item] = p.getTile(curr.item.x, curr.item.y)
				i += p.getInten(curr.item.x, curr.item.y)
				b, i = findBlock(p, curr.item.x, curr.item.y, c, b, i)
			}
			curr = curr.next
		}
	}
	return b, i
}
