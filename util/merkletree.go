package util

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	eth "github.com/ethereum/go-ethereum/crypto"
)

type TreeContent interface {
	CalculateHash() ([]byte, error)
	Equals(other TreeContent) (bool, error)
}

type MerkleTree struct {
	Root         *Node
	merkleRoot   []byte
	Leafs        []*Node
	hashStrategy func() hash.Hash
	sort         bool
}

type Node struct {
	Tree   *MerkleTree
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	Hash   []byte
	C      TreeContent
	sort   bool
}

type DefaultCont struct {
	Data string
}

// func Test() {
// 	//组织参数
// 	db := database.GetDb()
// 	var wls []model.Wl
// 	db.Where("status = 1").Find(&wls)

// 	var contents []TreeContent

// 	// var contentData []byte

// 	var custLeaf DefaultCont
// 	var wl model.Wl

// 	for _, vwl := range wls {
// 		v := EncodePack(vwl.Addr, big.NewInt(vwl.MintPrice), big.NewInt(vwl.MintNum), big.NewInt(vwl.StartTime), big.NewInt(vwl.EndTime))

// 		// contentData = append(contentData, []byte(v+"\n")...)

// 		// fmt.Println(v)
// 		contents = append(contents, DefaultCont{
// 			Data: v,
// 		})

// 		if strings.EqualFold(vwl.Addr, "0xa46262167990c81Cd66060067A6C107D7EbF01A2") {
// 			// leafT := eth.Keccak256Hash(hexutil.MustDecode(v)).Hex()
// 			// fmt.Println("aa: ", leafT)

// 			wl = vwl
// 			custLeaf = DefaultCont{Data: v}
// 		}
// 	}

// 	tree, _ := NewTreeWithHashStrategySorted(contents, sha3.NewLegacyKeccak256, true)

// 	fmt.Println("merkleRoot: ", hexutil.Encode(tree.merkleRoot))

// 	merklePath, index, err := tree.GetMerklePathHex(custLeaf)

// 	fmt.Println(merklePath, index, err)

// 	b, _ := tree.VerifyContent(custLeaf)

// 	fmt.Println("b===========:", b)

// 	wl.Remark = ""

// 	// fmt.Println(merklePath, index, err)

// 	// a2tmp, err := decimal.NewFromString("6666664497663742508244")

// 	// leaf := encodePack("0x39f0357140C66629E3640cc767F6F18b4A4D6a33", a2tmp.BigInt())
// 	// leafT := eth.Keccak256Hash(hexutil.MustDecode(leaf)).Hex()
// 	// fmt.Println(leafT)

// 	// some := DefaultCont{Data: leaf}
// 	// s1, s2, err := tree.GetMerklePathHex(some)
// 	// fmt.Println(s1, s2, err)

// }

// func Test2() {
// 	//组织参数
// 	db := database.GetDb()
// 	var wls []model.Wl
// 	db.Where("status = 1").Find(&wls)

// 	var contents []TreeContent

// 	// var contentData []byte
// 	var custLeafs []DefaultCont
// 	// var wl model.Wl

// 	for _, vwl := range wls {
// 		v := EncodePack(vwl.Addr, big.NewInt(vwl.MintPrice), big.NewInt(vwl.MintNum), big.NewInt(vwl.StartTime), big.NewInt(vwl.EndTime))

// 		// contentData = append(contentData, []byte(v+"\n")...)

// 		// fmt.Println(v)
// 		contents = append(contents, DefaultCont{
// 			Data: v,
// 		})

// 		custLeaf := DefaultCont{Data: v}
// 		custLeafs = append(custLeafs, custLeaf)

// 	}

// 	tree, _ := NewTreeWithHashStrategySorted(contents, sha3.NewLegacyKeccak256, true)

// 	fmt.Println("merkleRoot: ", hexutil.Encode(tree.merkleRoot))

// 	for i, vwl := range wls {

// 		if vwl.Remark == "新Vega" || vwl.Remark == "新Vega1500" || vwl.Remark == "校友卡" || vwl.Remark == "新Vega179" {
// 			merklePath, _, _ := tree.GetMerklePathHex(custLeafs[i])
// 			// fmt.Println(merklePath, index, err)
// 			a := strings.Join(merklePath, ",")

// 			db.Exec("update vg_mm t set t.proof = ? where UPPER(t.address) = upper(?)", a, vwl.Addr)
// 		}

// 	}

// 	// fmt.Println(merklePath, index, err)

// 	// a2tmp, err := decimal.NewFromString("6666664497663742508244")

// 	// leaf := encodePack("0x39f0357140C66629E3640cc767F6F18b4A4D6a33", a2tmp.BigInt())
// 	// leafT := eth.Keccak256Hash(hexutil.MustDecode(leaf)).Hex()
// 	// fmt.Println(leafT)

// 	// some := DefaultCont{Data: leaf}
// 	// s1, s2, err := tree.GetMerklePathHex(some)
// 	// fmt.Println(s1, s2, err)

// }

// 钱包地址+铸造单价+最大数量+开始时间+结束时间
func EncodePack(address string, mintPrice *big.Int, mintNum *big.Int, startTime *big.Int, endTime *big.Int) string {
	a1 := hexutil.MustDecode(address)
	a2 := mintPrice.Bytes()
	a3 := mintNum.Bytes()
	a4 := startTime.Bytes()
	a5 := endTime.Bytes()

	var a11, a21, a31, a41, a51 [32]byte
	for i := range a1 {
		a11[32-len(a1)+i] = a1[i]
	}

	for i := range a2 {
		a21[32-len(a2)+i] = a2[i]
	}

	for i := range a3 {
		a31[32-len(a3)+i] = a3[i]
	}

	for i := range a4 {
		a41[32-len(a4)+i] = a4[i]
	}

	for i := range a5 {
		a51[32-len(a5)+i] = a5[i]
	}

	var splitdata []byte
	for _, v := range a11 {
		splitdata = append(splitdata, v)
	}

	for _, v := range a21 {
		splitdata = append(splitdata, v)
	}

	for _, v := range a31 {
		splitdata = append(splitdata, v)
	}

	for _, v := range a41 {
		splitdata = append(splitdata, v)
	}

	for _, v := range a51 {
		splitdata = append(splitdata, v)
	}

	// leaf := eth.Keccak256Hash(splitdata).Hex()
	// if address == "0xFaD652ad544a85c87F00285EB64acDa7d79E20B6" {
	// 	fmt.Println("break")
	// }
	// return leaf
	return hexutil.Encode(splitdata)
}

func EncodePackAirdorp(address string, mintPrice *big.Int) string {
	a1 := hexutil.MustDecode(address)
	a2 := mintPrice.Bytes()
	var a11, a21 [32]byte
	for i := range a1 {
		a11[32-len(a1)+i] = a1[i]
	}

	for i := range a2 {
		a21[32-len(a2)+i] = a2[i]
	}

	var splitdata []byte
	for _, v := range a11 {
		splitdata = append(splitdata, v)
	}

	for _, v := range a21 {
		splitdata = append(splitdata, v)
	}

	// leaf := eth.Keccak256Hash(splitdata).Hex()
	// if address == "0xFaD652ad544a85c87F00285EB64acDa7d79E20B6" {
	// 	fmt.Println("break")
	// }
	// return leaf
	return hexutil.Encode(splitdata)
}

func (t DefaultCont) CalculateHash() ([]byte, error) {
	return eth.Keccak256(hexutil.MustDecode(t.Data)), nil
}

func (t DefaultCont) Equals(other TreeContent) (bool, error) {
	return t.Data == other.(DefaultCont).Data, nil
}

func sortAppend(sort bool, a, b []byte) []byte {
	if !sort {
		return append(a, b...)
	}
	var aBig, bBig big.Int
	aBig.SetBytes(a)
	bBig.SetBytes(b)
	if aBig.Cmp(&bBig) == -1 {
		return append(a, b...)
	}
	return append(b, a...)
}

func (n *Node) verifyNode(sort bool) ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}
	rightBytes, err := n.Right.verifyNode(sort)
	if err != nil {
		return nil, err
	}

	leftBytes, err := n.Left.verifyNode(sort)
	if err != nil {
		return nil, err
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(sortAppend(sort, leftBytes, rightBytes)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (n *Node) calculateNodeHash(sort bool) ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(sortAppend(sort, n.Left.Hash, n.Right.Hash)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func NewTree(cs []TreeContent) (*MerkleTree, error) {
	var defaultHashStrategy = sha256.New
	t := &MerkleTree{
		hashStrategy: defaultHashStrategy,
		sort:         false,
	}
	root, leafs, err := buildWithContent(cs, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}

func NewTreeWithHashStrategy(cs []TreeContent, hashStrategy func() hash.Hash) (*MerkleTree, error) {
	t := &MerkleTree{
		hashStrategy: hashStrategy,
		sort:         false,
	}
	root, leafs, err := buildWithContent(cs, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}

func NewTreeWithHashStrategySorted(cs []TreeContent, hashStrategy func() hash.Hash, sort bool) (*MerkleTree, error) {
	t := &MerkleTree{
		hashStrategy: hashStrategy,
		sort:         sort,
	}
	root, leafs, err := buildWithContent(cs, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}

func (m *MerkleTree) GetMerklePath(content TreeContent) ([][]byte, []int64, error) {
	for _, current := range m.Leafs {
		ok, err := current.C.Equals(content)
		if err != nil {
			return nil, nil, err
		}

		if ok {
			currentParent := current.Parent
			var merklePath [][]byte
			var index []int64
			for currentParent != nil {
				if !currentParent.dup {
					if bytes.Equal(currentParent.Left.Hash, current.Hash) {
						merklePath = append(merklePath, currentParent.Right.Hash)
						index = append(index, 1) // right leaf
					} else {
						merklePath = append(merklePath, currentParent.Left.Hash)
						index = append(index, 0) // left leaf
					}
				}
				current = currentParent
				currentParent = currentParent.Parent
			}
			return merklePath, index, nil
		}
	}
	return nil, nil, nil
}

func (m *MerkleTree) GetMerklePathHex(content TreeContent) ([]string, []int64, error) {
	for _, current := range m.Leafs {
		ok, err := current.C.Equals(content)
		if err != nil {
			return nil, nil, err
		}

		if ok {
			currentParent := current.Parent
			var merklePath []string
			var index []int64
			for currentParent != nil {
				if !currentParent.dup {
					if bytes.Equal(currentParent.Left.Hash, current.Hash) {
						merklePath = append(merklePath, hexutil.Encode(currentParent.Right.Hash))
						index = append(index, 1) // right leaf
					} else {
						merklePath = append(merklePath, hexutil.Encode(currentParent.Left.Hash))
						index = append(index, 0) // left leaf
					}
				}
				current = currentParent
				currentParent = currentParent.Parent
			}
			return merklePath, index, nil
		}
	}
	return nil, nil, nil
}

func buildWithContent(cs []TreeContent, t *MerkleTree) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, errors.New("error: cannot construct tree with no content")
	}
	var leafs []*Node
	for _, c := range cs {
		hash, err := c.CalculateHash()
		if err != nil {
			return nil, nil, err
		}

		leafs = append(leafs, &Node{
			Hash: hash,
			C:    c,
			leaf: true,
			Tree: t,
		})
	}

	root, err := buildIntermediate(leafs, t)
	if err != nil {
		return nil, nil, err
	}

	return root, leafs, nil
}

func buildIntermediate(nl []*Node, t *MerkleTree) (*Node, error) {
	var nodes []*Node
	rangelen := len(nl)
	if len(nl)%2 == 1 {
		rangelen = rangelen - 1
	}
	for i := 0; i < rangelen; i += 2 {
		h := t.hashStrategy()
		var left, right int = i, i + 1
		// if i+1 == len(nl) {
		// 	right = i
		// }
		chash := sortAppend(t.sort, nl[left].Hash, nl[right].Hash)
		if _, err := h.Write(chash); err != nil {
			return nil, err
		}
		n := &Node{
			Left:  nl[left],
			Right: nl[right],
			Hash:  h.Sum(nil),
			Tree:  t,
		}
		nodes = append(nodes, n)
		nl[left].Parent = n
		nl[right].Parent = n
		if len(nl) == 2 {
			return n, nil
		}
	}
	if len(nl)%2 == 1 {
		duplicate := &Node{
			Hash: nl[len(nl)-1].Hash,
			C:    nl[len(nl)-1].C,
			leaf: true,
			dup:  true,
			Tree: t,
		}
		nodes = append(nodes, duplicate)
		nl[len(nl)-1].Parent = duplicate
	}

	return buildIntermediate(nodes, t)
}

func (m *MerkleTree) MerkleRoot() []byte {
	return m.merkleRoot
}

func (m *MerkleTree) RebuildTree() error {
	var cs []TreeContent
	for _, c := range m.Leafs {
		cs = append(cs, c.C)
	}
	root, leafs, err := buildWithContent(cs, m)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

func (m *MerkleTree) RebuildTreeWith(cs []TreeContent) error {
	root, leafs, err := buildWithContent(cs, m)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

func (m *MerkleTree) VerifyTree() (bool, error) {
	calculatedMerkleRoot, err := m.Root.verifyNode(m.sort)
	if err != nil {
		return false, err
	}

	if bytes.Compare(m.merkleRoot, calculatedMerkleRoot) == 0 {
		return true, nil
	}
	return false, nil
}

func (m *MerkleTree) VerifyContent(content TreeContent) (bool, error) {
	for _, l := range m.Leafs {
		ok, err := l.C.Equals(content)
		if err != nil {
			return false, err
		}

		if ok {
			currentParent := l.Parent
			for currentParent != nil {
				h := m.hashStrategy()
				rightBytes, err := currentParent.Right.calculateNodeHash(m.sort)
				if err != nil {
					return false, err
				}

				leftBytes, err := currentParent.Left.calculateNodeHash(m.sort)
				if err != nil {
					return false, err
				}

				if _, err := h.Write(sortAppend(m.sort, leftBytes, rightBytes)); err != nil {
					return false, err
				}
				if bytes.Compare(h.Sum(nil), currentParent.Hash) != 0 {
					return false, nil
				}
				currentParent = currentParent.Parent
			}
			return true, nil
		}
	}
	return false, nil
}

func (n *Node) String() string {
	return fmt.Sprintf("%t %t %v %s", n.leaf, n.dup, n.Hash, n.C)
}

func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprint(l)
		s += "\n"
	}
	return s
}
