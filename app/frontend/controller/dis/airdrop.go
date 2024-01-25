package dis

import (
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"eli/app/model"
	"eli/database"
	ml "eli/middleware"
	"eli/util"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/sha3"
)

// 检查白单
func CheckWhite(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")
	address := c.PostForm("address")

	if len(address) != 42 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100014"))
		return
	}
	db := database.DB

	var airdrop model.DisAirdrop

	db.Where(" address = ? ", address).First(&airdrop)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"airdrop": airdrop}))
}

func GetMakerProof(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")
	address := c.PostForm("address")

	if len(address) != 42 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100014"))
		return
	}
	db := database.DB

	var airdrop model.DisAirdrop

	var airdrops []model.DisAirdrop

	db.Where(" status = 1 ").Find(&airdrops)

	var contents []util.TreeContent

	// var contentData []byte

	var custLeaf util.DefaultCont

	for _, air := range airdrops {

		v := util.EncodePackAirdorp(air.Address, big.NewInt(int64(air.Amount)))

		// contentData = append(contentData, []byte(v+"\n")...)

		// fmt.Println(v)
		contents = append(contents, util.DefaultCont{
			Data: v,
		})

		if strings.EqualFold(air.Address, address) {
			// leafT := eth.Keccak256Hash(hexutil.MustDecode(v)).Hex()
			// fmt.Println("aa: ", leafT)

			airdrop = air
			custLeaf = util.DefaultCont{Data: v}
		}
	}

	tree, _ := util.NewTreeWithHashStrategySorted(contents, sha3.NewLegacyKeccak256, true)

	fmt.Println("merkleRoot: ", hexutil.Encode(tree.MerkleRoot()))

	merklePath, index, err := tree.GetMerklePathHex(custLeaf)

	fmt.Println(merklePath, index, err)

	b, _ := tree.VerifyContent(custLeaf)

	fmt.Println("b===========:", b)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"airdrop": airdrop, "proof": merklePath}))

	// fmt.Println(merklePath, index, err)

	// a2tmp, err := decimal.NewFromString("6666664497663742508244")

	// leaf := encodePack("0x39f0357140C66629E3640cc767F6F18b4A4D6a33", a2tmp.BigInt())
	// leafT := eth.Keccak256Hash(hexutil.MustDecode(leaf)).Hex()
	// fmt.Println(leafT)

	// some := DefaultCont{Data: leaf}
	// s1, s2, err := tree.GetMerklePathHex(some)
	// fmt.Println(s1, s2, err)

}
