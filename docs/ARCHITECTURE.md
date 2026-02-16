# Finsplitter - Architecture Description

## 1. Overview

**Finsplitter** is a financial expense management and splitting system designed to help users track credit card expenses and split costs among multiple people. The system supports both personal expense tracking and collaborative expense sharing among family members, friends, or roommates.

### 1.1 Purpose

The primary goals of Finsplitter are:

1. **Expense Tracking**: Allow users to track credit card transactions and monthly bills
2. **Expense Splitting**: Enable splitting of expenses among multiple people with flexible percentage rules
3. **Settlement Calculation**: Calculate how much each person owes to settle shared expenses
4. **Shared Finances Management**: Support managing shared household or group finances

### 1.2 Target Users

-   Individuals tracking personal expenses across multiple credit cards
-   Couples or families sharing credit card expenses
-   Roommates splitting household purchases
-   Friends sharing recurring subscriptions or group expenses

---

## 2. Core Concepts

### 2.1 Domain Entities

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FINSPLITTER DOMAIN                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌──────────┐         ┌──────────┐         ┌─────────────┐                 │
│   │   User   │────────▶│  Person  │◀────────│  Card Brand │                 │
│   └──────────┘         └──────────┘         └─────────────┘                 │
│        │                    │                      │                        │
│        │ owns               │ participates         │ categorizes            │
│        ▼                    ▼                      ▼                        │
│   ┌──────────┐         ┌──────────┐         ┌──────────┐                    │
│   │   Card   │◀────────│Card-Person│────────▶│   Card   │                   │
│   └──────────┘         └──────────┘         └──────────┘                    │
│        │                                          │                         │
│        │ generates                                │ contains                │
│        ▼                                          ▼                         │
│   ┌──────────┐                              ┌─────────────┐                 │
│   │   Bill   │◀─────────────────────────────│ Transaction │                 │
│   └──────────┘         groups               └─────────────┘                 │
│                        ▲                       │                         │
│                        │                       │ splits                  │
│                        │ many-to-many          ▼                         │
│                     ┌──────────────┐    ┌───────────────────┐              │
│                     │ Transaction  │    │ Transaction-Person │             │
│                     │     -Bill    │    └───────────────────┘              │
│                     └──────────────┘                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### User

A registered system user who can:

-   Own and manage credit cards
-   Create and manage people (participants for expense splitting)
-   Enter and import transactions
-   View settlement summaries

#### Person

Represents an individual who participates in expense splitting. A person:

-   May or may not be a registered User in the system
-   Can be associated with one or more Cards (to define default split percentages)
-   Can be assigned portions of specific Transactions

#### Card Brand

Categorizes credit cards by issuer (e.g., Visa, Mastercard, American Express).

#### Card

Represents a credit card belonging to a User. Contains:

-   Card identification (name, last 4 digits, brand)
-   Billing cycle information (due date, closing date)
-   Card tier (e.g., Black, Platinum)
-   Default sharing rules with People

#### Bill

Represents a monthly credit card statement. Contains:

-   Reference period (month/year)
-   Due date information
-   Payment status and date

#### Transaction

An individual purchase or charge on a credit card. Contains:

-   Transaction details (identifier, name, value, date)
-   Recurring charge flag (for automatic propagation)
-   Installment information
-   Association with one or more Bills via a pivot table (for installment distribution)

---

## 3. Business Rules

### 3.1 Transaction Management

| Rule ID | Rule                   | Description                                                                                                                               |
| ------- | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| TXN-001 | Non-zero values        | Transaction values must be non-zero. Positive values represent debits (purchases), negative values represent credits (refunds, cashback). |
| TXN-002 | Installment validation | If a transaction has installments, the number must be positive (> 0). Maximum of 18 installments.                                                                      |
| TXN-003 | Installment distribution  | Transactions with installments are associated with multiple bills via a pivot table. Each bill contains the transaction value divided by the installments number. |
| TXN-004 | Installment update     | Updating installments_number deletes all transaction_bill records and re-creates them based on the new installment count.                                        |
| TXN-005 | Transaction deletion   | Deleting a transaction cascades to all related transaction_bill records (ON DELETE CASCADE).                                                                         |

### 3.2 Bill Management

| Rule ID  | Rule                   | Description                                                                                                 |
| -------- | ---------------------- | ----------------------------------------------------------------------------------------------------------- |
| BILL-001 | Unique per card/period | Each card can have only one bill per month/year combination.                                                |
| BILL-002 | Date validation        | Month must be 1-12, year must be reasonable (2000-2100), due dates must be 1-31.                            |
| BILL-003 | Payment consistency    | `paid_on` date can only be set when `paid = true`.                                                          |
| BILL-004 | Closing date behavior  | Transactions on or after closing_date go to next month bill. Transactions before closing_date go to current month bill. |
| BILL-005 | Bill date replication  | Bill due_date and due_month are copied from card's due_date at bill creation time (preserves history if card dates change). |
| BILL-006 | Future bill creation   | Installment transactions proactively create all future bill records upon transaction creation.              |
| BILL-007 | Transaction history    | Transactions cannot be created more than 18 months in the past.                                             |

### 3.3 Expense Splitting

| Rule ID   | Rule                       | Description                                                                                                    |
| --------- | -------------------------- | -------------------------------------------------------------------------------------------------------------- |
| SPLIT-001 | Percentage bounds          | Split percentages must be between 0% and 100%.                                                                 |
| SPLIT-002 | Card-level defaults        | A Person associated with a Card has a `default_percentage` that applies to all transactions unless overridden. |
| SPLIT-003 | Transaction-level override | Individual transactions can have custom split percentages that override card-level defaults.                   |
| SPLIT-004 | Historical tracking        | Changes to sharing arrangements are tracked via `end_date` timestamps, preserving history.                     |

### 3.4 Alternating Payment (Planned)

| Rule ID | Rule                  | Description                                                                                                             |
| ------- | --------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| ALT-001 | Rotation schedule     | Recurring transactions can be configured for alternating payments where participants take turns paying the full amount. |
| ALT-002 | Rotation cycle        | The rotation follows a defined order (e.g., Person 1 → Person 2 → Person 1 → ...).                                      |
| ALT-003 | Settlement adjustment | Settlement calculations account for whose "turn" it was to pay.                                                         |

---

## 4. Key Workflows

### 4.1 Manual Transaction Entry

```
┌─────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────────┐
│  User   │────▶│ Select Card │────▶│ Enter Txn   │────▶│ Define Split    │
└─────────┘     └─────────────┘     │   Details   │     │   (optional)    │
                                     └─────────────┘     └─────────────────┘
                                                                │
                                                                ▼
                                     ┌─────────────┐     ┌─────────────────┐
                                     │ System Auto │────▶│ Transaction     │
                                     │ Assigns Bill│     │ Created & Linked│
                                     │             │     │ to Bill(s)       │
                                     └─────────────┘     └─────────────────┘
```

**Steps:**

1. User selects the credit card for the transaction
2. User enters transaction details (name, value, date, identifier, installments if applicable)
3. User optionally defines how the transaction should be split
4. System automatically assigns the transaction to the appropriate bill(s) based on:
    - Transaction date
    - Card's closing date (determines which billing cycle)
    - Installment number (creates links to future bills for each installment)
5. Transaction is created with calculated split values and linked to bills via transaction_bill pivot table

### 4.2 Transaction Import

```
┌─────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────────┐
│  User   │────▶│ Upload File │────▶│ Parse &     │────▶│ Review &        │
└─────────┘     │ (CSV/OFX)   │     │ Match       │     │ Confirm         │
                └─────────────┘     └─────────────┘     └─────────────────┘
                                                               │
                                                               ▼
                                    ┌─────────────┐     ┌─────────────────┐
                                    │ Transactions│────▶│ User Defines    │
                                    │ Imported    │     │ Splits          │
                                    └─────────────┘     └─────────────────┘
```

**Steps:**

1. User uploads a statement file (CSV, OFX, or other supported format)
2. System parses the file and attempts to match with existing transactions, if no match is found, the transaction is later created as a new transaction.
3. User reviews imported transactions and makes adjustments
4. Transactions are created in bulk (without splits)
5. User manually defines splits for each transaction

### 4.3 Recurring Transaction Propagation

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ New Bill    │────▶│ Find        │────▶│ Link        │────▶│ Apply Split │
│ Created     │     │ Transactions│     │ Transaction │     │ Rules       │
└─────────────┘     │ from Card   │     │ to Bill     │     └─────────────┘
                    └─────────────┘     │ (Pivot)      │
                                        └─────────────┘
```

**Steps:**

1. When a new bill is created for a card
2. System finds all applicable transactions for that bill based on:
   - Transactions marked as `recurring_charge = true`
   - Installment transactions that have remaining installments
3. System creates links between transactions and the new bill via the transaction_bill pivot table
4. Each bill receives the transaction value divided by the number of installments
5. Split rules are applied (including alternating payment rotation if configured)

### 4.4 Settlement Calculation

```
┌─────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────────┐
│  User   │────▶│ Select      │────▶│ Calculate   │────▶│ Display         │
└─────────┘     │ Period/Bill │     │ Balances    │     │ Settlement      │
                └─────────────┘     └─────────────┘     └─────────────────┘
```

**Settlement Logic:**

1. For each transaction in the selected period/bill:
    - Identify who paid (card owner)
    - Calculate the installment amount (transaction value / number of installments) if applicable
    - Identify the split (transaction_person records)
    - Calculate each person's share for this installment
2. Aggregate amounts by person pairs
3. Display net settlement amounts ("Person A owes Person B: $X")

**Example:**

```
Bill Total: $1,000 (Card owned by Alice)
Split: Alice 60%, Bob 40%

- Alice paid: $1,000
- Alice's share: $600
- Bob's share: $400

Settlement: Bob owes Alice $400

Example with installments:
Transaction: $300 over 3 installments (Card owned by Alice)
Bill 1 contains 1st installment: $100
Split: Alice 60%, Bob 40%

- Alice paid: $100
- Alice's share: $60
- Bob's share: $40

Settlement for Bill 1: Bob owes Alice $40
```

### 4.5 Alternating Payment Workflow (Planned)

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Recurring   │────▶│ Check       │────▶│ Assign Full │────▶│ Update      │
│ Transaction │     │ Rotation    │     │ Amount to   │     │ Rotation    │
│ Copied      │     │ Schedule    │     │ Current     │     │ Counter     │
└─────────────┘     └─────────────┘     │ Person      │     └─────────────┘
                                        └─────────────┘
```

**Example - Netflix Subscription Alternating:**

```
Month 1: Alice pays 100% ($15)
Month 2: Bob pays 100% ($15)
Month 3: Alice pays 100% ($15)
...

Settlement after 3 months:
- Alice paid: $30
- Bob paid: $15
- Fair share each: $22.50

Bob owes Alice: $7.50
```

---

## 5. System Architecture

### 5.1 Technology Stack

| Layer             | Technology                                    |
| ----------------- | --------------------------------------------- |
| Language          | Go                                            |
| Database          | PostgreSQL                                    |
| API Style         | REST (OpenAPI)                                |
| Container         | Docker                                        |
| Query Builder     | SQLC (type-safe SQL) & Squirrel (SQL builder) |
| API Framework     | Huma v2                                       |
| HTTP Framework    | Chi                                           |
| Testing Framework | Testify                                       |
| Logging           | slog/OpenTelemetry                            |
| Tracing           | OpenTelemetry                                 |
| Metrics           | OpenTelemetry                                 |

### 5.2 Project Structure

```
finsplitter/
├── api/                    # OpenAPI specifications
├── cmd/
│   └── api/               # Application entrypoint
├── internal/
│   ├── app/
│   │   ├── ports/         # Repository interfaces
│   │   └── usecases/      # Business logic (by domain)
│   ├── config/            # Configuration management
│   ├── domain/
│   │   ├── entity/        # Domain entities
│   │   └── errs/          # Domain errors
│   └── gateways/
│       ├── http/          # HTTP handlers and routing
│       └── postgres/      # Database implementation
│           ├── migrations/# Database migrations
│           └── sqlc/      # Generated query code
└── pkg/
    └── telemetry/         # Observability (logging, tracing, metrics)
```

### 5.3 Architecture Pattern

The project follows a **Hexagonal Architecture** (Ports and Adapters):

```
                    ┌─────────────────────────────────────────┐
                    │              HTTP Gateway               │
                    │         (internal/gateways/http)        │
                    └─────────────────┬───────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────────┐
                    │              Use Cases                  │
                    │         (internal/app/usecases)         │
                    │                                         │
                    │  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
                    │  │  Card   │ │  Bill   │ │  Txn    │   │
                    │  │  Brand  │ │         │ │         │   │
                    │  └─────────┘ └─────────┘ └─────────┘   │
                    └─────────────────┬───────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────────┐
                    │               Ports                     │
                    │          (internal/app/ports)           │
                    │                                         │
                    │         Repository Interfaces           │
                    └─────────────────┬───────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────────┐
                    │          PostgreSQL Gateway             │
                    │       (internal/gateways/postgres)      │
                    └─────────────────────────────────────────┘
```

---

## 6. Database Schema

### 6.1 Schema Overview

The database follows a relational model with pivot tables for many-to-many relationships:

- **transaction_bill**: Links transactions to bills (for installment distribution)
- **transaction_person**: Links transactions to people (for expense splitting)
- **card_person**: Links cards to people (for default split percentages)

### 6.2 Transaction-Bill Pivot Table

The `transaction_bill` table links transactions to bills for installment distribution:

```
┌─────────────────────────────────────────────┐
│           transaction_bill                   │
├─────────────────────────────────────────────┤
│ transaction_id (FK) │ bill_id (FK)         │
├─────────────────────────────────────────────┤
│ uuid                 │ uuid                 │
│ ─────────────────    │ ─────────────────    │
│ Links to transaction │ Links to bill        │
└─────────────────────────────────────────────┘
```

**Example - Transaction with 3 installments:**

```
Transaction: $900 (3 installments of $300 each)
Card: Visa (closing date: 15th, due date: 5th)

transaction_bill records:
┌─────────────────────────┬──────────────────────┐
│ transaction_id          │ bill_id              │
├─────────────────────────┼──────────────────────┤
│ abc-123 (original txn)  │ bill-1 (Jan 2025)    │
│ abc-123 (original txn)  │ bill-2 (Feb 2025)    │
│ abc-123 (original txn)  │ bill-3 (Mar 2025)    │
└─────────────────────────┴──────────────────────┘

Each bill shows: $300 installment
```

### 6.3 Key Constraints

| Table             | Constraint                    | Description                         |
| ----------------- | ----------------------------- | ----------------------------------- |
| user              | UNIQUE(lower(email))          | Case-insensitive unique emails      |
| user              | UNIQUE(lower(username))       | Case-insensitive unique usernames   |
| card_brand        | UNIQUE(lower(name))           | Case-insensitive unique brand names |
| bill              | UNIQUE(card_id, year, month)  | One bill per card per month         |
| card.l4d          | CHECK(l4d ~ '^[0-9]{4}$')     | Exactly 4 numeric digits            |
| transaction.value | CHECK(value != 0)             | Non-zero transaction values         |
| \*\_percentage    | CHECK(0 <= percentage <= 100) | Valid percentage range              |

---

### 7. Future Considerations

| Feature             | Description                                             | Priority |
| ------------------- | ------------------------------------------------------- | -------- |
| Web app             | Web application for managing expenses                   | High     |
| Import formats      | Support CSV, OFX, and bank-specific formats             | High     |
| Payment tracking    | Track actual payments made between people               | Medium   |
| Notifications       | Remind users of upcoming due dates                      | Medium   |
| Reports             | Monthly/yearly expense reports and charts               | Medium   |
| Multi-currency      | Support transactions in different currencies            | Low      |
| Mobile app          | Native mobile application                               | Low      |
| Alternating Payment | Support alternating payments for recurring transactions | Low      |

---

## 8. Glossary

| Term                 | Definition                                                                            |
| -------------------- | ------------------------------------------------------------------------------------- |
| **Bill**             | A monthly credit card statement containing transactions from a billing period         |
| **Card**             | A credit card belonging to a user                                                     |
| **Card Person**      | Association between a card and a person, defining default split percentages           |
| **Person**           | An individual who participates in expense splitting (may or may not be a system user) |
| **Recurring Charge** | A transaction that repeats monthly (e.g., subscriptions)                              |
| **Settlement**       | The calculated amount one person owes another after accounting for shared expenses    |
| **Split**            | The division of a transaction's cost among multiple people                            |
| **Transaction**      | An individual purchase, charge, or credit on a credit card                            |
| **Transaction-Bill** | Association between a transaction and one or more bills for installment distribution |
| **User**             | A registered account in the system                                                    |
