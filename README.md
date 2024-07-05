# kubectl-thunder

This kubelet plugin implement high level features for kubectl.

## Features

### Node high level commands

- [x] `kubectl thunder nodes verify-allocation` - Check if a node can allocate a given resource request.

```
OPTIONS:
   --memory value                               Memory resource to try to allocate
   --cpu value                                  CPU resource to try to allocate
   --show-labels value [ --show-labels value ]  Print node labels
   --help, -h                                   show help
```

### Secrets high level commands

- [x] `kubectl thunder secrets find expiring ` - Find secrets that are about to expire.

```
OPTIONS:
   --used-only    Only report certificates that are used by ingresses (default: false)
   --unused-only  Only report certificates that are not used by ingresses (default: false)
   --delete       Delete expired certificates after reporting them (use with caution!) (default: false)
   --after value  Only consider certificates that are expired after the specified number of days (default: 0)
   --help, -h     show help
```

- [x] `kubectl thunder secrets find by-certificate-san` - Find secrets that are not used by any ingresses.

```
OPTIONS:
   --san value       Subject Alternative Name to search for
   --wildcard-match  Subject Alternative Name is matched by a wildcard certificate (default: false)
   --help, -h        show help
```