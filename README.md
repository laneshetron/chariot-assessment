```
+-------------------------------------------------------------------------------+
|                                   ID FORMAT                                   |
+-------------------------------------------------------------------------------+
|    Seconds since Jan 1, 2020    | Dash |   Counter   |      Random Data       |
+---------------------------------+------+-------------+------------------------+
| 32 bits                         | 1 bit| 16 bits     | 40 bits                |
+---------------------------------+------+-------------+------------------------+
| BC5YUGA                         |  -   | AA          | TBBDEINA               |
+-------------------------------------------------------------------------------+
|                89 bits Total (Base32 Encoded to 19 characters)                |
+-------------------------------------------------------------------------------+
|                       FORMAT:  XXXXXXX-XXXXXXXXXXX                            |
+-------------------------------------------------------------------------------+
```

# EXAMPLE ID

Generated ID: BC5YUGA-AATBBDEINA

- `BC5YUGA` maps to the first 48 bits (32 bits for the timestamp + 16 bits for the counter).
- `TBBDEINA` maps to the last 40 bits (random data).

# DESCRIPTION

1. **Seconds since Jan 1, 2020 (32 bits):**
   - Represents the number of seconds since January 1, 2020 (UTC).
   - Provides temporal context and helps ensure IDs are time-ordered.
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

# ENCODING

- **Base32 Encoding:**
  - The entire 89-bit ID (including the dash) is encoded using Base32 without padding.
  - The output is 19 characters long.

- **Output Format:**
  - The encoded ID is split into two segments: `XXXXXXX-XXXXXXXXXXX`, separated by a dash.
  - The dash improves readability.

# NOTES

- The format ensures a balance between time-sequential ordering and randomness, making the IDs both predictable in terms of creation order and resistant to collisions.


# Benchmarks

## With cryptographically secure randomness (default)

```
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkGenerateID-8   	  739557	      1448 ns/op
PASS
ok  	github.com/laneshetron/chariot-assessment	1.552s
```

## Without secure randomness

```
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkGenerateID-8   	 4797810	       244.4 ns/op
PASS
ok  	github.com/laneshetron/chariot-assessment	1.880s
```
