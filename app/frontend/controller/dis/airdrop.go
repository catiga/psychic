package dis

import (
	"fmt"
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

	fmt.Println("address:", address)

	if len(address) != 42 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100014"))
		return
	}

	db := database.DB

	var airdrop model.DisAirdrop
	db.Where(" address = ? ", address).First(&airdrop)

	if airdrop == (model.DisAirdrop{}) {
		c.JSON(http.StatusOK, ml.Fail(lang, "100015"))
		return
	}

	var airdrops []model.DisAirdrop

	db.Where(" status = 1 ").Find(&airdrops)

	var contents []util.TreeContent

	// var contentData []byte

	var custLeaf util.DefaultCont

	for _, air := range airdrops {

		fmt.Println(air.Address, air.Amount.BigInt())

		v := util.EncodePackAirdorp(air.Address, air.Amount.BigInt())

		// contentData = append(contentData, []byte(v+"\n")...)

		// fmt.Println(v)
		contents = append(contents, util.DefaultCont{
			Data: v,
		})

		if strings.EqualFold(air.Address, address) {
			custLeaf = util.DefaultCont{Data: v}
		}
	}

	tree, _ := util.NewTreeWithHashStrategySorted(contents, sha3.NewLegacyKeccak256, true)

	fmt.Println("merkleRoot: ", hexutil.Encode(tree.MerkleRoot()))

	merklePath, index, err := tree.GetMerklePathHex(custLeaf)

	fmt.Println(merklePath, index, err)

	// b, _ := tree.VerifyContent(custLeaf)

	// fmt.Println("b===========:", b)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"airdrop": airdrop, "proof": merklePath}))
}
