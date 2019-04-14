package actor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func reverseString(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}
	return
}

func TestMessagesIntegrity(t *testing.T) {
	t.Log("FatalErrorMessage")
	{
		assert.Equal(t, "TO FROM EE", FatalErrorMessage("FROM", "TO"))
	}

	t.Log("AccountCreatedMessage")
	{
		assert.Equal(t, "TO FROM AN", AccountCreatedMessage("FROM", "TO"))
	}

	t.Log("PromiseAcceptedMessage")
	{
		assert.Equal(t, "TO FROM X0", PromiseAcceptedMessage("FROM", "TO"))
	}

	t.Log("CommitAcceptedMessage")
	{
		assert.Equal(t, "TO FROM X1", CommitAcceptedMessage("FROM", "TO"))
	}

	t.Log("RollbackAcceptedMessage")
	{
		assert.Equal(t, "TO FROM X2", RollbackAcceptedMessage("FROM", "TO"))
	}

	t.Log("AccountStateMessage")
	{
		assert.Equal(t, "TO FROM S0 CURRENCY t BALANCE PROMISED", AccountStateMessage("FROM", "TO", "CURRENCY", "BALANCE", "PROMISED", true))
		assert.Equal(t, "TO FROM S0 CURRENCY f BALANCE PROMISED", AccountStateMessage("FROM", "TO", "CURRENCY", "BALANCE", "PROMISED", false))
	}
}

func TestSymetricityOfEvents(t *testing.T) {
	t.Log("FatalError (* -> " + FatalError + ")")
	{
		kind := strings.Split(FatalErrorMessage("FROM", "TO"), " ")[2]
		assert.Equal(t, FatalError, kind)
		assert.Equal(t, kind, reverseString(kind))
	}

	t.Log("New Account -> (" + ReqCreateAccount + " -> " + reverseString(ReqCreateAccount) + ")")
	{
		kind := strings.Split(AccountCreatedMessage("FROM", "TO"), " ")[2]
		assert.Equal(t, ReqCreateAccount, reverseString(kind))
	}

	t.Log("Promise -> (" + PromiseOrder + " -> " + reverseString(PromiseOrder) + ")")
	{
		kind := strings.Split(PromiseAcceptedMessage("FROM", "TO"), " ")[2]
		assert.Equal(t, PromiseOrder, reverseString(kind))
	}

	t.Log("Commit -> (" + CommitOrder + " -> " + reverseString(CommitOrder) + ")")
	{
		kind := strings.Split(CommitAcceptedMessage("FROM", "TO"), " ")[2]
		assert.Equal(t, CommitOrder, reverseString(kind))
	}

	t.Log("Rollback -> (" + RollbackOrder + " -> " + reverseString(RollbackOrder) + ")")
	{
		kind := strings.Split(RollbackAcceptedMessage("FROM", "TO"), " ")[2]
		assert.Equal(t, RollbackOrder, reverseString(kind))
	}
}
