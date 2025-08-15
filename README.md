# ds

Dead simple Git repository manager. Fast status checks across multiple accounts.

## Install

```bash
git clone https://github.com/verlyn13/ds-go.git
cd ds-go
make install
```

## Usage

```bash
ds status         # show all repos
ds status -d      # show dirty repos only  
ds fetch          # update remote info
ds scan           # rebuild index
```

## Config

Creates `~/.config/ds/config.yaml` on first run:

```yaml
base_dir: ~/Projects

accounts:
  verlyn13:
    type: personal
    ssh_host: github-personal
  jjohnson-47:
    type: school
    ssh_host: github-work
```

## Build

```bash
make build        # build binary
make test         # run tests
make install      # install to /usr/local/bin
```

MIT License