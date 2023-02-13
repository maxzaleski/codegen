# codegen

## Presence

Config is contained within root `.codegen` directory:

- Global config is labeled as `config.yaml`
- Seperate pkg configs can be defined via `pkg` > `pkg/name.yaml`
- Seperate template can be defined and specified via `tmpl` > `tmpl/layer_addon.tmpl`

## Steps of Implementation

1. Parse various config files into `Spec`
2. Check for presence: if a pkg has already been generated, skip
3. Generate

## Extensions

- Http