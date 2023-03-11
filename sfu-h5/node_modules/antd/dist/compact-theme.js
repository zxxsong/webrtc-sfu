const { compactThemeSingle } = require('./theme');
const defaultTheme = require('./default-theme');

module.exports = {
  ...compactThemeSingle,
  ...defaultTheme
}