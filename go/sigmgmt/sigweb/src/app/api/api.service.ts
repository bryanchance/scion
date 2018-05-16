import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'

import { CIDR, DefaultSession, IA, PathSelector, Policy, Session, SIG, Site } from '../sites/models'
import { User } from './user.service'

@Injectable()
export class ApiService {

  constructor(
    private http: HttpClient,
  ) { }

  /**
   * Sites
   */
  getSites() {
    return this.http.get<Site[]>('sites')
  }

  getSite(name: string) {
    return this.http.get<Site>('sites/' + name)
  }

  createSite(site: Site) {
    return this.http.post<Site>('sites', site)
  }

  updateSite(site: Site) {
    return this.http.put<Site>('sites/' + site.Name, site)
  }

  deleteSite(site: Site) {
    return this.http.delete('sites/' + site.Name)
  }

  reloadConfig(name: string) {
    return this.http.get('sites/' + name + '/reload-config')
  }

  /**
   * Path predicates / selectors
   */
  getPathSelectors(site: Site) {
    return this.http.get<PathSelector[]>('sites/' + site.Name + '/paths')
  }

  createPathSelector(site: Site, ps: PathSelector) {
    return this.http.post<PathSelector>('sites/' + site.Name + '/paths', ps)
  }

  updatePathSelector(site: Site, ps: PathSelector) {
    return this.http.put('sites/' + site.Name + '/paths/' + ps.Name, ps)
  }

  deletePathSelector(site: Site, ps: PathSelector) {
    return this.http.delete('sites/' + site.Name + '/paths/' + ps.Name)
  }

  /**
   * Remote ASes
   */
  getIAs(site: Site) {
    return this.http.get<IA[]>('sites/' + site.Name + '/ias')
  }

  getIA(site: Site, ia: IA) {
    return this.http.get<IA>(this.iaUrl(site, ia))
  }

  createIA(site: Site, ia: IA) {
    return this.http.post<IA>('sites/' + site.Name + '/ias', ia)
  }

  updateIA(site: Site, ia: IA) {
    return this.http.put<string>(this.iaUrl(site, ia), ia)
  }

  updateIAPolicies(site: Site, ia: IA, policy: string) {
    return this.http.put<string>(this.iaUrl(site, ia) + '/policies', { Policy: policy })
  }

  deleteIA(site: Site, ia: IA) {
    return this.http.delete(this.iaUrl(site, ia))
  }

  /**
   * AS Entries
   */
  /** Networks */
  getNetworks(site: Site, ia: IA) {
    return this.http.get<CIDR[]>(this.iaUrl(site, ia) + '/networks')
  }

  createNetwork(site: Site, ia: IA, network: CIDR) {
    return this.http.post<CIDR>(this.iaUrl(site, ia) + '/networks', network)
  }

  deleteNetwork(site: Site, ia: IA, network: CIDR) {
    return this.http.delete(this.iaUrl(site, ia) + '/networks/' + network.ID)
  }

  /** SIGS */
  getSIGs(site: Site, ia: IA) {
    return this.http.get<SIG[]>(this.iaUrl(site, ia) + '/sigs')
  }

  createSIG(site: Site, ia: IA, sig: SIG) {
    return this.http.post<SIG>(this.iaUrl(site, ia) + '/sigs', sig)
  }

  updateSIG(site: Site, ia: IA, sig: SIG) {
    return this.http.put<SIG>(this.iaUrl(site, ia) + '/sigs/' + sig.ID, sig)
  }

  deleteSIG(site: Site, ia: IA, sig: SIG) {
    return this.http.delete(this.iaUrl(site, ia) + '/sigs/' + sig.ID)
  }

  /** Sessions */
  getSessions(site: Site, ia: IA) {
    return this.http.get<Session[]>(this.iaUrl(site, ia) + '/sessions')
  }

  createSession(site: Site, ia: IA, session: Session) {
    return this.http.post<Session>(this.iaUrl(site, ia) + '/sessions', session)
  }

  deleteSession(site: Site, ia: IA, session: Session) {
    return this.http.delete(this.iaUrl(site, ia) + '/sessions/' + session.ID)
  }

  getDefaultSession(site: Site, ia: IA) {
    return this.http.get<DefaultSession>(this.iaUrl(site, ia) + '/session-default')
  }

  setDefaultSession(site: Site, ia: IA, defaultSession: DefaultSession) {
    return this.http.post(this.iaUrl(site, ia) + '/session-default', defaultSession)
  }

  getSessionAliases(site: Site, ia: IA) {
    return this.http.get<string[]>(this.iaUrl(site, ia) + '/session-aliases')
  }

  /** Authentication */
  obtainToken(user: User) {
    return this.http.post('auth', user)
  }

  /**
   * Combine site and ia to a string
   * @param site
   * @param ia
   * @returns {string}
   */
  private iaUrl(site: Site, ia: IA): string {
    return 'sites/' + site.Name + '/ias/' + ia.ISD + '-' + ia.AS
  }
}
