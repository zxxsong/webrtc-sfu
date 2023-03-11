# react-mdi

Material-UI community Material Design Icons.

[Community Material Design Icons](https://materialdesignicons.com/) as svg react components, built with [icon-builder](https://github.com/gabriel-miranda/material-ui/tree/master/icon-builder) from Material-UI.

Special thanks to [Austin Andrews](https://github.com/Templarian) for managing Material Design Icons.

## Installation

```sh
npm install react-mdi
```

## Usage

```js
import React from 'react';

import AccountIcon from 'react-mdi/icons/account';

export default class Account extends React.Component {
  render() {
    return (
      <AccountIcon size={16} className="myClassName" />
    );
  }
}
```

## Props
| Prop        | Default value | Usage                                                                                                  |
|:------------|:--------------|:-------------------------------------------------------------------------------------------------------|
| `size`      | **24**        | Used to set the `height`and `width` in the style attribute                                             |
| `className` | **null**      | Used to apply a css class to the component and properly style it                                       |

## Build

```sh
npm run build
```
