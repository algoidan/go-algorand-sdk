package test

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
	"io/ioutil"
	"strings"

	"github.com/cucumber/godog"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorand/go-algorand-sdk/encoding/msgpack"
)

func iBuildAnApplicationTransactionUnit(
	operation string,
	applicationIdInt int,
	sender, approvalProgram, clearProgram string,
	globalBytes, globalInts, localBytes, localInts int,
	appArgs, foreignApps, foreignAssets, appAccounts string,
	fee, firstValid, lastValid int,
	genesisHash string, extraPages int) error {

	applicationId = uint64(applicationIdInt)
	var clearP []byte
	var approvalP []byte
	var err error

	if approvalProgram != "" {
		approvalP, err = ioutil.ReadFile("features/resources/" + approvalProgram)
		if err != nil {
			return err
		}
	}

	if clearProgram != "" {
		clearP, err = ioutil.ReadFile("features/resources/" + clearProgram)
		if err != nil {
			return err
		}
	}
	args, err := parseAppArgs(appArgs)
	if err != nil {
		return err
	}
	var accs []string
	if appAccounts != "" {
		accs = strings.Split(appAccounts, ",")
	}

	fApp, err := splitUint64(foreignApps)
	if err != nil {
		return err
	}

	fAssets, err := splitUint64(foreignAssets)
	if err != nil {
		return err
	}

	gSchema := types.StateSchema{NumUint: uint64(globalInts), NumByteSlice: uint64(globalBytes)}
	lSchema := types.StateSchema{NumUint: uint64(localInts), NumByteSlice: uint64(localBytes)}

	suggestedParams, err := getSuggestedParams(uint64(fee), uint64(firstValid), uint64(lastValid), "", genesisHash, true)
	if err != nil {
		return err
	}

	switch operation {
	case "create":
		if extraPages > 0 {
			tx, err = future.MakeApplicationCreateTxWithExtraPages(false, approvalP, clearP,
				gSchema, lSchema, args, accs, fApp, fAssets,
				suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{}, uint32(extraPages))
		} else {
			tx, err = future.MakeApplicationCreateTx(false, approvalP, clearP,
				gSchema, lSchema, args, accs, fApp, fAssets,
				suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{})
		}

		if err != nil {
			return err
		}

	case "update":
		tx, err = future.MakeApplicationUpdateTx(applicationId, args, accs, fApp, fAssets,
			approvalP, clearP,
			suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{})
		if err != nil {
			return err
		}

	case "call":
		tx, err = future.MakeApplicationCallTx(applicationId, args, accs,
			fApp, fAssets, types.NoOpOC, approvalP, clearP, gSchema, lSchema,
			suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{})
	case "optin":
		tx, err = future.MakeApplicationOptInTx(applicationId, args, accs, fApp, fAssets,
			suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{})
		if err != nil {
			return err
		}

	case "clear":
		tx, err = future.MakeApplicationClearStateTx(applicationId, args, accs, fApp, fAssets,
			suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{})
		if err != nil {
			return err
		}

	case "closeout":
		tx, err = future.MakeApplicationCloseOutTx(applicationId, args, accs, fApp, fAssets,
			suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{})
		if err != nil {
			return err
		}

	case "delete":
		tx, err = future.MakeApplicationDeleteTx(applicationId, args, accs, fApp, fAssets,
			suggestedParams, addr1, nil, types.Digest{}, [32]byte{}, types.Address{})
		if err != nil {
			return err
		}
	}
	return nil

}

func feeFieldIsInTxn() error {
	var txn map[string]interface{}
	err := msgpack.Decode(stx, &txn)
	if err != nil {
		return fmt.Errorf("Error while decoding txn. %v", err)
	}
	if _, ok := txn["txn"].(map[interface{}]interface{})["fee"]; !ok {
		return fmt.Errorf("fee field missing. %v", err)
	}
	return nil
}

func feeFieldNotInTxn() error {
	var txn map[string]interface{}
	err := msgpack.Decode(stx, &txn)
	if err != nil {
		return fmt.Errorf("Error while decoding txn. %v", err)
	}
	if _, ok := txn["txn"].(map[interface{}]interface{})["fee"]; ok {
		return fmt.Errorf("fee field found but it should have been omitted. %v", err)
	}
	return nil
}

func theBaseEncodedSignedTransactionShouldEqual(base int, golden string) error {
	gold, err := base64.StdEncoding.DecodeString(golden)
	if err != nil {
		return err
	}
	stxStr := base64.StdEncoding.EncodeToString(stx)
	if !bytes.Equal(gold, stx) {
		return fmt.Errorf("Application signed transaction does not match the golden: %s != %s", stxStr, golden)
	}
	return nil
}

func getSuggestedParams(
	fee, fv, lv uint64,
	gen, ghb64 string,
	flat bool) (types.SuggestedParams, error) {
	gh, err := base64.StdEncoding.DecodeString(ghb64)
	if err != nil {
		return types.SuggestedParams{}, err
	}
	return types.SuggestedParams{
		Fee:             types.MicroAlgos(fee),
		GenesisID:       gen,
		GenesisHash:     gh,
		FirstRoundValid: types.Round(fv),
		LastRoundValid:  types.Round(lv),
		FlatFee:         flat,
	}, err
}

func weMakeAGetAssetByIDCall(assetID int) error {
	clt, err := algod.MakeClient(mockServer.URL, "")
	if err != nil {
		return err
	}
	clt.GetAssetByID(uint64(assetID)).Do(context.Background())
	return nil
}

func weMakeAGetApplicationByIDCall(applicationID int) error {
	clt, err := algod.MakeClient(mockServer.URL, "")
	if err != nil {
		return err
	}
	clt.GetApplicationByID(uint64(applicationID)).Do(context.Background())
	return nil
}

func weMakeASearchForApplicationsCall(applicationID int) error {
	clt, err := indexer.MakeClient(mockServer.URL, "")
	if err != nil {
		return err
	}
	clt.SearchForApplications().ApplicationId(uint64(applicationID)).Do(context.Background())
	return nil
}

func weMakeALookupApplicationsCall(applicationID int) error {
	clt, err := indexer.MakeClient(mockServer.URL, "")
	if err != nil {
		return err
	}
	clt.LookupApplicationByID(uint64(applicationID)).Do(context.Background())
	return nil
}

func ApplicationsUnitContext(s *godog.Suite) {
	// @unit.transactions
	s.Step(`^fee field is in txn$`, feeFieldIsInTxn)
	s.Step(`^fee field not in txn$`, feeFieldNotInTxn)

	//@unit.applications
	s.Step(`^we make a GetAssetByID call for assetID (\d+)$`, weMakeAGetAssetByIDCall)
	s.Step(`^we make a GetApplicationByID call for applicationID (\d+)$`, weMakeAGetApplicationByIDCall)
	s.Step(`^we make a SearchForApplications call with applicationID (\d+)$`, weMakeASearchForApplicationsCall)
	s.Step(`^we make a LookupApplications call with applicationID (\d+)$`, weMakeALookupApplicationsCall)
}
