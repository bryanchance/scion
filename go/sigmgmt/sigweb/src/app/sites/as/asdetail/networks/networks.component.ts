import { Component, Input, OnChanges, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'

import { ApiService } from '../../../../api/api.service'
import { CIDR, IA, Site } from '../../../models'

@Component({
  selector: 'ana-networks',
  templateUrl: './networks.component.html',
  styleUrls: ['./networks.component.scss']
})
export class NetworksComponent implements OnChanges {
  @Input() site: Site
  @Input() ia: IA
  success = ''
  error = ''

  networks: CIDR[]
  network = new CIDR
  @ViewChild('networkForm') form: NgForm

  constructor(private api: ApiService) { }

  ngOnChanges(): void {
    if (this.site.Name && this.ia) {
      this.api.getNetworks(this.site, this.ia).subscribe(
        networks => this.networks = networks
      )
    }
  }

  onSubmit() {
    this.error = ''
    this.api.createNetwork(this.site, this.ia, this.network).subscribe(
      network => {
        this.networks.push({ ...network })
        this.form.resetForm()
      },
      error => this.error = error
    )
  }

  delete(idx: number) {
    this.api.deleteNetwork(this.site, this.ia, this.networks[idx]).subscribe(
      () => this.networks.splice(idx, 1)
    )
  }

}
