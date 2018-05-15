import { AppPage } from './app.po'

describe('sigweb App', () => {
  let page: AppPage

  beforeEach(() => {
    page = new AppPage()
  })

  it('should display menu title', () => {
    page.navigateTo()
    expect(page.getAppTitle()).toEqual('SIGWeb: Policy Configurator')
  })
})
