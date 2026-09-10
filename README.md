# ExploreTech Full Stack Interview

-feedback Breaking down the commits and the frontend components into more manageable pieces would have gone a long way in facilitating effective communication and collaboration, which is critical in this role. 
possible one PR-sized idea per commit; one visible UI concern per file.
--Thank you for taking the time to interview with ExploreTech!
Please don't spend more than 2 hours on this take-home exercise.

Try to push your submission 24 hours before your live interview
so that the team can review it prior to the interview.

Be prepared to discuss what you built during the live interview.

This is a starter project that is representative of our tech stack:

- Go backend
- React frontend

## Prompt

In this exercise, you are tasked with building a tool to help a geophysicist perform 
quality control on geophysics data. We often get big, complicated datasets from field sensors 
and it is painful to sift through it to make sure the data was collected properly. 

This exercise is designed to illuminate how you approach building tools for an open-ended problem.
There is no right answer. You do not need to know any geophysics. You just need to 
understand the problem and build a convenient way for a non-engineer to perform a task.

We encourage the use of AI tools, but even the best tools have limitations. 
In the live interview, we will discuss together how you used AI and your own problem-solving skills
together to address the prompt.

More information can be found in the "Your Task" section.

## Grading

You will be evaluated mainly on your product sense (what should be built and why) 
and your software design (is it clean and organized or messy).
There's more here than most can finish in 2 hours, so focus on demonstrating your
approach rather than aiming for a 100% complete solution. 

## Questions

Clarifying questions are welcome and encouraged. To ask a question, please send 
written questions to Kai Wells (kai.wells@exploretech.ai) or book a brief phone call (see email).

## Submission

To submit, you can push your code changes directly to the main branch on this repository.

## Project Setup

Before you begin, please ensure that you are able to run this project.

### Docker / Podman

There is a single Dockerfile which builds the backend and frontend and ships the
build artifacts in a minimal container.

```bash
docker build --platform=linux/amd64 -t et-interview .
docker run --rm --platform=linux/amd64 -p 8080:8080 et-interview
```

You should see a bar chart rendered at [localhost:8080](http://localhost:8080).

### Native

For more rapid development, it's recommended that you run these components natively.

```bash
cd backend
go run .
```

```bash
cd frontend
npm install
npm run dev
```

The frontend dev server is configured to proxy API requests to `localhost:8080`.

You should see a bar chart rendered at [localhost:5173](http://localhost:5173).

### Tests

```bash
cd backend && go test ./...
cd frontend && npm test
```

There are a few tests on each side already. They are there to save you the setup,
not to set a bar - write as many or as few of your own as you think the work
warrants.

### What's already wired

- **SQLite Database.** The schema in `backend/internal/db/schema.sql` is applied
  at startup and before every import; add your tables there. Statements must be
  idempotent, since it runs every time the process starts.
- **Persistence.** The database file is `backend/qc.db` by default, so data
  survives a restart. Override with `DB_CONN`, and delete the file if you want
  a clean slate.
- **A throwaway vertical slice.** `GET /api/demo` goes handler → store →
  SQLite → JSON and feeds the bar chart on the frontend. It exists to prove the
  wiring works end to end. Delete `backend/internal/demo`, the `demo_values`
  table, and the route once you're oriented.

## AI Coding Agents

You are encouraged to use the AI coding agent/harness of your choice when working on
this project. You will be provided with an OpenRouter API key with $50 in credits.

## Your Task

Derrick is a geophysicist reviewing DCIP survey data before it is used in an inversion.
Last week we collected data from 4 lines near where we expect to find an anomaly in
the subsurface. Most readings are fine, but field work is inherently messy so some of
the data is junk (bad electrode contact, instrument glitches, duplicate readings, etc.).
If junk gets through, the inversion results could be trashed as we try to match our
physics simulations to physically impossible data.

Build a tool to help Derrick perform quality control.

**You do not need to know any geophysics.** The columns are described in the
data section below.

### Level 1

- Import survey data from CSV into the backend database with validation (CLI)
- List a line's measurements (API and UI)
- Record a verdict for each measurement (pass, fail, unreviewed) (API and UI)

### Level 2

- Filtering and sorting that makes triage manageable (UI, optional API)
- A visual representation of the data (UI)
- Bulk verdicts in a single request (API and UI)

### Level 3

- Group repeated locations and surface disagreement between measurements
- Summary of review progress
- Undo
- Decay curve inspection

## The Data

Four CSV files in `data/`, one per survey line, **7,507 readings total**:

| File | Readings |
|---|---|
| `data/line-a.csv` | 1,888 |
| `data/line-b.csv` | 1,875 |
| `data/line-c.csv` | 1,886 |
| `data/line-d.csv` | 1,858 |

Import is a CLI subcommand rather than a browser upload - we'd rather see the
parse/validate/persist path than HTTP form plumbing:

```bash
cd backend
go run . import ../data/line-a.csv
go run . import ../data/line-*.csv
```

`backend/import.go` opens the file and leaves the rest to you.

If you prefer to implement the data loading and analysis in another language,
feel free to do so. Please commit any scripts/tools to your submission.
Make sure that your import script and the Go backend agree on the database schema.

### Columns

30 columns, one row per reading.

**Identity**

| Column | Meaning |
|---|---|
| `id` | Reading identifier assigned by the acquisition system, e.g. `line-a-r00001`. |
| `line_id` | Which line the reading belongs to: `line-a` ... `line-d`. |

**Electrodes.** Four per reading: a transmitting pair (`tx1`, `tx2`) that
injects current, and a receiving pair (`rx1`, `rx2`) that measures the voltage
it produces. Each has four columns - `<slot>_id`, `_lat`, `_lon`, `_alt`.

| Column | Meaning |
|---|---|
| `tx1_id`, `tx2_id`, `rx1_id`, `rx2_id` | Electrode identifier, e.g. `line-a-e17` - the line plus a 1-based position along it. **Stable for the whole survey:** a given physical electrode keeps its id, so the same electrode appears in many readings. |
| `*_lat`, `*_lon` | WGS84 decimal degrees. |
| `*_alt` | Meters. Constant 180 m here - the ground is flat. |

**Acquisition**

| Column | Meaning |
|---|---|
| `timestamp` | RFC 3339 UTC. Real acquisition order - roughly one line per day, 2026-08-17 to 2026-08-20. |
| `array_type` | Electrode arrangement. Always `dipole-dipole` in this data. |
| `symmetry` | The dipole-dipole **n factor**, 1-8: how far the receiving pair sits from the transmitting pair, in multiples of the electrode spacing. Larger n reaches deeper and returns a weaker signal. |
| `stacks` | **The number of current pulses averaged together within this one reading** (3-8). This is *not* how many times the geometry was measured - see "Repeats" below. Each row is already one stacked reading. |
| `input_current` | Current injected by the transmitting pair, in amperes. |
| `contact_resistance` | Ohms. A property of the **transmitting pair** on this reading - how well those two electrodes are coupled to the ground. It says nothing about the row's receiving pair. |

**Measurements**

| Column | Meaning |
|---|---|
| `apparent_resistivity` | Ω·m. |
| `apparent_resistivity_err` | Ω·m. The instrument's own estimate of the uncertainty on that value. |
| `chargeability` | V/V. How much the ground holds charge after the current is switched off. |
| `chargeability_err` | V/V. As above. |
| `decay_curve` | 10 semicolon-separated V/V values: the voltage decay sampled in 10 time gates after current shutoff, centred at 24.32, 35.96, 53.18, 78.64, 116.30, 171.97, 254.31, 376.06, 556.10 and 822.34 ms. Normally decays toward zero. |

**Review**

| Column | Meaning |
|---|---|
| `qc_review` | Present and empty in every file. This is the column your tool fills in. |

**Repeats.** The four electrodes of a reading define its *geometry*. Each
geometry is measured several times over - 6 to 8 readings each, 1,072 distinct
geometries across the 7,507 rows. Repeats of the same geometry should agree
with each other.
