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

# EXAMPLE ID

Generated ID: BC5YUGA-AATBBDEINA

- `BC5YUGA` maps to the first 32 bits (representing seconds since Jan 1, 2020)
- `AA` maps to the next 16 bits (monotonic intra-second counter)
- `TBBDEINA` maps to the last 40 bits (random data).

# DESCRIPTION

1. **Seconds since Jan 1, 2020 (32 bits):**
   - Represents the number of seconds since January 1, 2020 (UTC).
   - Provides temporal context and helps ensure IDs are time-ordered.
   - By adjusting the timestamp to a reference date of Jan 1, 2020 we can extend support by 50 years to 2156.
   - Example (encoded in Base32): `BC5YUGA`.

2. **Dash (`-`):**
   - A static dash character inserted after the timestamp and counter.
   - It improves readability by separating the two main segments of the ID.
   - Represented by a single bit in the ID format.

3. **Counter (16 bits):**
   - A counter that increments with each ID generated within the same second.
   - Resets to zero when the timestamp changes.
   - Example (encoded in Base32): `AA`.

4. **Random Data (40 bits):**
   - Cryptographically secure random bits, adding uniqueness and preventing collisions.
   - Example (encoded in Base32): `TBBDEINA`.

## Unique
- **Each ID is guaranteed unique:**
  - Within each second the 16-bit counter is incremented monotonically to guarantee order, followed by 40 random bits.
  - Assuming two separate ID generators generate within the same second with the same counter value, the odds of a collision are 1 in ~1.2 septillion.

## Random
- 40 cryptographically secure random bits are appended to each ID using the `crypto/rand` standard library.
- The use of a secure source of randomness does come at the cost of some performance (as noted below), but using any pseudorandom generator would render the IDs guessable by a motivated adversary.

## Human Readable
- Each ID is encoded in Base32 for case insensitivity.
- Each ID is lexicographically monotonic.
- A dash is added between the encoded timestamp and counter for readability.

## Sortable

## Compact
- Each ID is 18 characters in length, including the dash.

# ENCODING

- **Base32 Encoding:**
  - The entire 88-bit ID is encoded using Base32 without padding.
  - A dash is added after the encoded timestamp for readability
  - The output is 18 characters long.

# DISCUSSION

- The format ensures a balance between time-sequential ordering and randomness, making the IDs both predictable in terms of creation order and resistant to collisions.
- Ideally the counter would be 24-bits as opposed to 16 in order to ensure monotonicity at high rates within a single machine. Ultimately I chose not to do this:
  - Base32 encoding encodes 5 bits at a time, meaning that adding another byte (totaling 96 bits) would overflow to 20 characters. This would eliminate space for the dash.
  - 1 bit could be removed from either the random data or from the shifted timestamp without sacrificing precision, but this would require a custom encoder which seems beyond the scope of this assignment.
  - Base64 would have mapped all 96 bits to exactly 16 characters, but come at the cost of readability/case-insensitivity.
  - Even under heavy load, it is unlikely that ID generation on a single machine would exceed 65k/sec in a practical use case.
- The ID format does not guarantee monotonicity between machines.
  - Given the space constraints and the listed requirements, I chose to omit this consideration in favor of other priorities.


# Benchmarks

## With cryptographically secure randomness (default)

```
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkNew-8   	  953506	      1223 ns/op	      16 B/op	       1 allocs/op
PASS
ok  	github.com/laneshetron/chariot-assessment/pkg/id	1.635s
```

## Without secure randomness

```
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkNew-8   	14346354	        82.76 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/laneshetron/chariot-assessment/pkg/id	1.739s
```
