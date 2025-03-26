# copygen

Copygen is a tool that automates adding license headers to source files.

## Installing

From source:
```sh
git clone git@github.com:bschaatsbergen/copygen
cd copygen
make
```

## Usage

Create a `.copygen.yml` file in the root of your project with the following format:

```yaml
Header: |
  Copyright (c) 2024 Acme Inc.

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction...

Exclude:
  - "docs/**"
```

And run the below command to add the license headers to all the files in the project:

```sh
copygen .
```

If you prefer to see the changes before they are applied, you can run the below command:

```sh
$ ./copygen --dry-run .
```

## Contributing

Contributions are highly appreciated and always welcome.
Have a look through existing [Issues](https://github.com/bschaatsbergen/copygen/issues) and [Pull Requests](https://github.com/bschaatsbergen/copygen/pulls) that you could help with.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
