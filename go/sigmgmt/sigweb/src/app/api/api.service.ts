import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { map } from 'rxjs/operators'

import { ASEntry, CIDR, Site, TrafficPolicy, PathPolicyFile, SiteNetwork } from '../sites/models/models'
import { TrafficClass, TrafficClassFromJSON } from '../sites/models/models'
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
   * Path Policies
   */
  getPathPolicies() {
    return this.http.get<PathPolicyFile[]>('pathpolicies')
  }

  getPathPolicy(policy: string) {
    return this.http.get<PathPolicyFile>('pathpolicies/' + policy)
  }

  createPathPolicy(pp: PathPolicyFile) {
    return this.http.post<PathPolicyFile>('pathpolicies', pp)
  }

  updatePathPolicy(pp: PathPolicyFile) {
    return this.http.put<PathPolicyFile>('pathpolicies/' + pp.ID, pp)
  }

  deletePathPolicy(pp: PathPolicyFile) {
    return this.http.delete('pathpolicies/' + pp.ID)
  }

  validatePathPolicy(pp: PathPolicyFile) {
    return this.http.post<PathPolicyFile>('pathpolicies/validate', pp)
  }

  /**
   * IP Allocations / Site networks
   */
  getIPAllocations(site: Site) {
    return this.http.get<SiteNetwork[]>('sites/' + site.ID + '/allocations')
  }

  createIPAllocation(site: Site, sn: SiteNetwork) {
    return this.http.post<SiteNetwork>('sites/' + site.ID + '/allocations', sn)
  }

  updateIPAllocation(sn: SiteNetwork) {
    return this.http.put<string>('allocations/' + sn.ID, sn)
  }

  deleteIPAllocation(sn: SiteNetwork) {
    return this.http.delete('allocations/' + sn.ID)
  }

  /**
   * Traffic Classes
   */
  getTrafficClasses(site: Site) {
    return this.http.get<TrafficClass[]>('sites/' + site.ID + '/classes').pipe(
      map(classes => classes.map(tc => TrafficClassFromJSON(tc)))
    )
  }

  getTrafficClass(cls: string) {
    return this.http.get<TrafficClass>('classes/' + cls).pipe(
      map(tc => TrafficClassFromJSON(tc))
    )
  }

  createTrafficClass(site: Site, tc: TrafficClass) {
    const rawTc = Object.assign(new TrafficClass, tc)
    if (!tc.CondStr || tc.CondStr === '') {
      rawTc.CondStr = tc.condString
    } else {
      rawTc.CondStr = tc.CondStr
    }
    delete rawTc.Cond
    return this.http.post<TrafficClass>('sites/' + site.ID + '/classes', rawTc)
  }

  updateTrafficClass(tc: TrafficClass) {
    const rawTc = Object.assign(new TrafficClass, tc)
    if (!tc.CondStr || tc.CondStr === '') {
      rawTc.CondStr = tc.condString
    } else {
      rawTc.CondStr = tc.CondStr
    }
    delete rawTc.Cond
    return this.http.put<TrafficClass>('classes/' + tc.ID, rawTc).pipe(
      map(ntc => TrafficClassFromJSON(ntc)))
  }

  deleteTrafficClass(site: Site, tc: TrafficClass) {
    return this.http.delete('classes/' + tc.ID)
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

  updateAS(as: ASEntry) {
    return this.http.put<string>('ases/' + as.ID, as)
  }

  updateASPolicies(as: ASEntry) {
    return this.http.put<string>('ases/' + as.ID + '/policies', as)
  }

  deleteAS(as: ASEntry) {
    return this.http.delete('ases/' + as.ID)
  }

  /** Policies */
  getTrafficPolicies(as: ASEntry) {
    return this.http.get<TrafficPolicy[]>('ases/' + as.ID + '/policies')
  }

  createPolicy(as: ASEntry, policy: TrafficPolicy) {
    return this.http.post<TrafficPolicy>('ases/' + as.ID + '/policies', policy)
  }

  updatePolicy(as: ASEntry, policy: TrafficPolicy) {
    return this.http.put<TrafficPolicy>('ases/' + as.ID + '/policies/' + policy.ID, policy)
  }

  deletePolicy(policy: TrafficPolicy) {
    return this.http.delete('policies/' + policy.ID)
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

  deleteNetwork(network: CIDR) {
    return this.http.delete('networks/' + network.ID)
  }

  /** Authentication */
  obtainToken(user: User) {
    return this.http.post('auth', user)
  }
}
