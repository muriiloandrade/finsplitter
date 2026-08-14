<!-- Context: project-intelligence/business | Priority: high | Version: 2.0 | Updated: 2026-08-13 -->

# Business Domain

**Purpose**: The "why" and the domain contract of Finsplitter — problem statement, target users, domain entities, business rules (rule-ID tables), and key workflows. This is the canonical source of truth for the business/domain spec (formerly `docs/ARCHITECTURE.md` §1–§4, §8).

## Problem Statement

Tracking credit card expenses and splitting costs among multiple people (couples, families, roommates, friends) is manual, error-prone, and lacks structure. Finsplitter provides expense tracking, flexible percentage-based splitting, and settlement calculation in one system.

**Target users**: individuals tracking personal expenses across cards · couples/families sharing card expenses · roommates splitting household purchases · friends sharing recurring subscriptions.

## Core Goals

1. **Expense Tracking** — track credit card transactions and monthly bills
2. **Expense Splitting** — split expenses among people with flexible percentage rules
3. **Settlement Calculation** — compute net amounts owed ("Person A owes Person B: $X")
4. **Shared Finances** — manage shared household/group finances

## Domain Entities

| Entity | Definition |
|--------|-----------|
| **User** | Registered system user; owns cards, creates people, enters transactions |
| **Person** | Participant in expense splitting (may or may not be a registered User) |
| **Card Brand** | Card issuer category (Visa, Mastercard, Amex) |
| **Card** | A credit card belonging to a User (name, last-4, brand, billing cycle, tier, default sharing rules) |
| **Bill** | Monthly credit card statement (period, due date, payment status) |
| **Transaction** | Purchase/charge/credit on a card; optional installments, recurring flag, split rules |
| **Transaction-Bill** | Pivot: links a transaction to one or more bills (installment distribution) |
| **Card-Person** | Pivot: card ↔ person with `default_percentage` for splitting |

**Relationships** (condensed): User owns Card · User creates Person · Person participates in Card-Person · Card generates Bill · Card contains Transaction · Transaction splits via Transaction-Person · Transaction groups to Bill via Transaction-Bill (many-to-many, installment distribution).

## Business Rules

### Transaction Management (TXN)

| ID | Rule | Description |
|----|------|-------------|
| TXN-001 | Non-zero values | Transaction values must be non-zero. Positive = debits (purchases), negative = credits (refunds, cashback). |
| TXN-002 | Installment validation | If a transaction has installments, the number must be positive (> 0). Maximum of 18 installments. |
| TXN-003 | Installment distribution | Installment transactions link to multiple bills via the pivot table. Each bill contains value ÷ installments count. |
| TXN-004 | Installment update | Updating `installments_number` deletes all `transaction_bill` records and re-creates them for the new count. |
| TXN-005 | Transaction deletion | Deleting a transaction cascades to all related `transaction_bill` records (ON DELETE CASCADE). |

### Bill Management (BILL)

| ID | Rule | Description |
|----|------|-------------|
| BILL-001 | Unique per card/period | Each card can have only one bill per month/year combination. |
| BILL-002 | Date validation | Month must be 1–12, year must be reasonable (2000–2100), due dates must be 1–31. |
| BILL-003 | Payment consistency | `paid_on` date can only be set when `paid = true`. |
| BILL-004 | Closing date behavior | Transactions on/after `closing_date` go to next month's bill; before `closing_date` go to current month's bill. |
| BILL-005 | Bill date replication | Bill `due_date`/`due_month` copied from card's at creation (preserves history if card dates change). |
| BILL-006 | Future bill creation | Installment transactions proactively create all future bill records on transaction creation. |
| BILL-007 | Transaction history | Transactions cannot be created more than 18 months in the past. |

### Expense Splitting (SPLIT)

| ID | Rule | Description |
|----|------|-------------|
| SPLIT-001 | Percentage bounds | Split percentages must be between 0% and 100%. |
| SPLIT-002 | Card-level defaults | A Person on a Card has a `default_percentage` applying to all transactions unless overridden. |
| SPLIT-003 | Transaction-level override | Individual transactions can have custom split percentages overriding card-level defaults. |
| SPLIT-004 | Historical tracking | Sharing changes tracked via `end_date` timestamps, preserving history. |

### Alternating Payment (ALT — planned)

| ID | Rule | Description |
|----|------|-------------|
| ALT-001 | Rotation schedule | Recurring transactions can alternate payments: participants take turns paying the full amount. |
| ALT-002 | Rotation cycle | Rotation follows a defined order (e.g., Person 1 → Person 2 → Person 1 → …). |
| ALT-003 | Settlement adjustment | Settlement calculations account for whose "turn" it was to pay. |

## Key Workflows

1. **Manual transaction entry**: select card → enter details → optional split → system auto-assigns bill(s) by date + card closing date (BILL-004) + installments (BILL-006) → transaction created & linked via `transaction_bill`.
2. **Transaction import**: upload CSV/OFX → parse & match (unmatched → new transaction) → review/confirm → bulk create (no splits) → user defines splits.
3. **Recurring propagation**: new bill created → find applicable transactions (`recurring_charge = true`, remaining installments) → link via `transaction_bill` pivot → apply split rules.
4. **Settlement calculation**: per period/bill — identify payer (card owner), installment amount (value ÷ installments), split (`transaction_person`), each person's share → aggregate by person pairs → "Person A owes Person B: $X".

**Settlement example**:

```
Bill Total: $1,000 (Alice's card), Split: Alice 60% / Bob 40%
Alice paid: $1,000 · Alice share: $600 · Bob share: $400
Settlement: Bob owes Alice $400

With installments: Transaction $300 over 3 installments (Alice's card)
Bill 1 contains 1st installment: $100 · Split 60/40
Alice paid: $100 · Alice share: $60 · Bob share: $40
Settlement for Bill 1: Bob owes Alice $40
```

**Alternating payment example (planned)**: Netflix $15/mo — Month 1 Alice pays 100%, Month 2 Bob pays 100%, Month 3 Alice… After 3 months: Alice paid $30, Bob paid $15, fair share $22.50 each → Bob owes Alice $7.50.

## Future Considerations

| Feature | Priority |
|---------|----------|
| Web app | High |
| Import formats (CSV, OFX, bank-specific) | High |
| Payment tracking, Notifications, Reports | Medium |
| Multi-currency, Mobile app, Alternating payment | Low |

## 📂 Codebase References

- **Entities**: `internal/domain/entity/` (implemented: `user.go`, `card_brand.go`, `user_claims.go`)
- **Domain errors**: `internal/domain/errs/errs.go`
- **Schema**: `internal/gateways/postgres/migrations/` (32 migrations — card/bill/transaction schema migrated; app layer pending)
- **Technical mapping**: `business-tech-bridge.md` · **History**: `decisions-log.md` · **Roadmap**: `living-notes.md`
