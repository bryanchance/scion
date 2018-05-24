import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'

import { CIDR, ASEntry, PathSelector, Policy, SIG, Site } from '../sites/models'
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

  getSite(site: string) {
    return this.http.get<Site>('sites/' + site)
  }

  createSite(site: Site) {
    return this.http.post<Site>('sites', site)
  }

  updateSite(site: Site) {
    return this.http.put<Site>('sites/' + site.ID, site)
  }

  deleteSite(site: Site) {
    return this.http.delete('sites/' + site.ID)
  }

  reloadConfig(id: number) {
    return this.http.get('sites/' + id + '/reload-config')
  }

  /**
   * Path predicates / selectors
   */
  getPathSelectors(site: Site) {
    return this.http.get<PathSelector[]>('sites/' + site.ID + '/paths')
  }


  createPathSelector(site: Site, ps: PathSelector) {
    return this.http.post<PathSelector>('sites/' + site.ID + '/paths', ps)
  }

  updatePathSelector(ps: PathSelector) {
    return this.http.put('paths/' + ps.ID, ps)
  }

  deletePathSelector(site: Site, ps: PathSelector) {
    return this.http.delete('paths/' + ps.ID)
  }

  /**
   * Remote ASes
   */
  getASes(site: Site) {
    return this.http.get<ASEntry[]>('sites/' + site.ID + '/ases')
  }

  getAS(as: number) {
    return this.http.get<ASEntry>('ases/' + as)
  }

  createAS(site: Site, as: ASEntry) {
    return this.http.post<ASEntry>('sites/' + site.ID + '/ases', as)
  }

  updateAS(site: Site, as: ASEntry) {
    return this.http.put<string>('ases/' + as.ID, as)
  }

  updateASPolicies(as: ASEntry) {
    return this.http.put<string>('ases/' + as.ID + '/policies', as)
  }

  deleteAS(site: Site, as: ASEntry) {
    return this.http.delete('ases/' + as.ID)
  }

  /**
   * AS Entries
   */
  /** Networks */
  getNetworks(as: ASEntry) {
    return this.http.get<CIDR[]>('ases/' + as.ID + '/networks')
  }

  createNetwork(as: ASEntry, network: CIDR) {
    return this.http.post<CIDR>('ases/' + as.ID + '/networks', network)
  }

  deleteNetwork(as: ASEntry, network: CIDR) {
    return this.http.delete('networks/' + network.ID)
  }

  /** SIGS */
  getSIGs(as: ASEntry) {
    return this.http.get<SIG[]>('ases/' + as.ID + '/sigs')
  }

  getDefaultSIG() {
    return this.http.get<SIG>('sigs/default')
  }

  createSIG(as: ASEntry, sig: SIG) {
    return this.http.post<SIG>('ases/' + as.ID + '/sigs', sig)
  }

  updateSIG(sig: SIG) {
    return this.http.put<SIG>('sigs/' + sig.ID, sig)
  }

  deleteSIG(sig: SIG) {
    return this.http.delete('sigs/' + sig.ID)
  }

  /** Authentication */
  obtainToken(user: User) {
    return this.http.post('auth', user)
  }
}
