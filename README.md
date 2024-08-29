# Chariot Technical Assessment

## Setup

Assuming you have `docker` and `docker-compose` installed, run the following to start the service:
```
docker-compose build
docker-compose up
```

I've included a Postman collection (`collection.json`) in the root of this repo for easy testing of the API.

### Running tests & benchmarks

```
cd pkg/id
go test -bench=. -benchmem
```

## Part 1: Unique Identifiers

```
+-------------------------------------------------------------------------------+
|                                   ID FORMAT                                   |
+-------------------------------------------------------------------------------+
|    Seconds since Jan 1, 2020    | Dash |   Counter   |      Random Data       |
+---------------------------------+------+-------------+------------------------+
| 32 bits                         |      | 16 bits     | 40 bits                |
+---------------------------------+------+-------------+------------------------+
| BC5YUGA                         |  -   | AA          | TBBDEINA               |
+-------------------------------------------------------------------------------+
|                88 bits Total (Base32 Encoded to 18 characters)                |
+-------------------------------------------------------------------------------+
|                       FORMAT:  XXXXXXX-XXXXXXXXXXX                            |
+-------------------------------------------------------------------------------+
```

### Technical Specification

1. **Seconds since Jan 1, 2020 (32 bits):**
   - Represents the number of seconds since January 1, 2020 (UTC).
   - Provides temporal context and helps ensure IDs are time-ordered.
   - By adjusting the timestamp to a reference date of Jan 1, 2020 we can extend support by 50 years to 2156.
   - Example (encoded in Base32): `BC5YUGA`.

2. **Dash (`-`):**
   - A static dash character inserted between the timestamp and counter.
   - Improves readability by separating the two main segments of the ID.

3. **Counter (16 bits):**
   - A counter that increments with each ID generated within the same second.
   - Resets to zero when the timestamp changes.
   - Example (encoded in Base32): `AA`.

4. **Random Data (40 bits):**
   - Cryptographically secure random bits, adding uniqueness and preventing collisions.
   - Example (encoded in Base32): `TBBDEINA`.

##### EXAMPLE ID

Generated ID: BC5YUGA-AATBBDEINA

- `BC5YUGA` maps to the first 32 bits (representing seconds since Jan 1, 2020)
- `AA` maps to the next 16 bits (monotonic intra-second counter)
- `TBBDEINA` maps to the last 40 bits (random data).

#### Unique
- **Each ID is guaranteed unique:**
  - Within each second the 16-bit counter is incremented monotonically to guarantee order, followed by 40 random bits.
  - Assuming two separate ID generators generate within the same second with the same counter value, the odds of a collision are 1 in ~1.2 septillion.

#### Random
- 40 cryptographically secure random bits are appended to each ID using the `crypto/rand` standard library.
- The use of a secure source of randomness does come at the cost of some performance (as noted below), but using any pseudorandom generator would render the IDs guessable by a motivated adversary.

#### Human Readable
- Each ID is encoded in Base32 for case insensitivity.
- Each ID is lexicographically monotonic.
- A dash is added between the encoded timestamp and counter for readability.

#### Sortable
- **The ID is lexicographically monotonic.**
- Because many IDs may be generated within the same second, a 16-bit counter is prepended to the random data to ensure monotonicity.
- Note: Monotonicity is not guaranteed between separate machines.

#### Compact
- Each ID is 18 characters in length, including the dash.

### DISCUSSION

- The format ensures a balance between time-sequential ordering and randomness, making the IDs both predictable in terms of creation order and resistant to collisions.
- Ideally the counter would be 24-bits as opposed to 16 in order to ensure monotonicity at high rates within a single machine. Ultimately I chose not to do this:
  - Base32 encoding encodes 5 bits at a time, meaning that adding another byte (totaling 96 bits) would overflow to 20 characters. This would eliminate space for the dash.
  - 1 bit could be removed from either the random data or from the shifted timestamp without sacrificing precision, but this would require a custom encoder which seems beyond the scope of this assignment.
  - Base64 would have mapped all 96 bits to exactly 16 characters, but come at the cost of readability/case-insensitivity.
  - Even under heavy load, it is unlikely that ID generation on a single machine would exceed 65k/sec in a practical use case.
- The ID format does not guarantee monotonicity between machines.
  - Given the space constraints and the listed requirements, I chose to omit this consideration in favor of other priorities.


### Benchmarks

#### With cryptographically secure randomness (default)

```
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkNew-8   	  953506	      1223 ns/op	      16 B/op	       1 allocs/op
PASS
ok  	github.com/laneshetron/chariot-assessment/pkg/id	1.635s
```

#### Without secure randomness

```
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkNew-8   	14346354	        82.76 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/laneshetron/chariot-assessment/pkg/id	1.739s
```

## Part 2: Backend

### Tables:
- users
- accounts
- transactions

### Endpoints:
- GET  /health
- GET  /transactions
- POST /users
- POST /accounts
- GET  /accounts/:id/balance
- POST /accounts/:id/withdraw
- POST /accounts/:id/deposit
- POST /accounts/:id/transfer

### Idempotency
 - I employed an **end-to-end design** approach to guarantee idempotency.
 - Deposit, withdraw, and transfer requests include a `idempotency_key` field which uniquely identify a client's transaction.
   - Note: the idempotency_key is not the actual transaction id.
 - This `idempotency_key` can be anything, but a well-behaved client will likely use UUIDv4.
 - Key uniqueness is enforced at the table-level via a composite unique constraint on (`idempotency_key`, `type`) fields.
 - If a request is retried or attempts to reuse a consumed `idempotency_key`, the API will yield a `200` response and quietly discard the transaction.

### Concurrency & Isolation
- All transactions (deposit, withdraw, transfer) are conducted with the highest isolation level (`serializable`) to prevent race conditions.
- The deposit and withdraw endpoints use implicit locking for account updates; however, transfer uses explicit locking in order to prevent deadlocks.

### Transactions Cursor
- The `GET /transactions` endpoint returns a cursor-paginated list of transactions using the monotonic PK.
  - Each request queries page+1 results in order to determine if there is a next page, and sets `nextCursor` to the transaction ID of the page+1'th result if it exists.
  - `nextCursor` is included in the response body, and can be used to fetch the next page of results.
