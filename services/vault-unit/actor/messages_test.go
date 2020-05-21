package actor

import (
	"testing"

	"github.com/jancajthaml-openbank/vault-unit/model"

	//system "github.com/jancajthaml-openbank/actor-system"
	money "gopkg.in/inf.v0"

	"github.com/stretchr/testify/assert"
)

func TestMessagesIntegrity(t *testing.T) {
	t.Log("FatalErrorMessage")
	{
		assert.Equal(t, "EE", FatalErrorMessage())
	}

	t.Log("AccountCreatedMessage")
	{
		assert.Equal(t, "AN", AccountCreatedMessage())
	}

	t.Log("PromiseAcceptedMessage")
	{
		assert.Equal(t, "P1", PromiseAcceptedMessage())
	}

	t.Log("PromiseRejectedMessage")
	{
		assert.Equal(t, "P2 REASON", PromiseRejectedMessage("REASON"))
	}

	t.Log("CommitAcceptedMessage")
	{
		assert.Equal(t, "C1", CommitAcceptedMessage())
	}

	t.Log("CommitRejectedMessage")
	{
		assert.Equal(t, "C2 REASON", CommitRejectedMessage("REASON"))
	}

	t.Log("RollbackAcceptedMessage")
	{
		assert.Equal(t, "R1", RollbackAcceptedMessage())
	}

	t.Log("RollbackRejectedMessage")
	{
		assert.Equal(t, "R2 REASON", RollbackRejectedMessage("REASON"))
	}

	t.Log("AccountStateMessage")
	{
		balance, _ := new(money.Dec).SetString("1.0")
		promised, _ := new(money.Dec).SetString("2.0")

		a := model.Account{
			Format:         "FORMAT",
			Currency:       "CURRENCY",
			IsBalanceCheck: true,
			Balance:        balance,
			Promised:       promised,
		}

		assert.Equal(t, "S0 FORMAT CURRENCY t 1.0 2.0", AccountStateMessage(a))

		b := model.Account{
			Format:         "FORMAT",
			Currency:       "CURRENCY",
			IsBalanceCheck: false,
			Balance:        balance,
			Promised:       promised,
		}

		assert.Equal(t, "S0 FORMAT CURRENCY f 1.0 2.0", AccountStateMessage(b))
	}
}
