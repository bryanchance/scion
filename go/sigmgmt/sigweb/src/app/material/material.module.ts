import { NgModule } from '@angular/core'
import {
  MatButtonModule,
  MatCardModule,
  MatExpansionModule,
  MatFormFieldModule,
  MatIconModule,
  MatInputModule,
  MatListModule,
  MatProgressSpinnerModule,
  MatSelectModule,
  MatToolbarModule,
} from '@angular/material'

@NgModule({
  imports: [
    MatToolbarModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatExpansionModule,
    MatListModule,
    MatIconModule,
    MatCardModule,
    MatSelectModule,
    MatProgressSpinnerModule
  ],
  declarations: [],
  exports: [
    MatToolbarModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatExpansionModule,
    MatListModule,
    MatIconModule,
    MatCardModule,
    MatSelectModule,
    MatProgressSpinnerModule
  ]
})
export class MaterialModule { }
