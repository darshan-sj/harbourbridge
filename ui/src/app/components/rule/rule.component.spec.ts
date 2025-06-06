import { ComponentFixture, TestBed } from '@angular/core/testing'

import { RuleComponent } from './rule.component'
import { MatTableModule } from '@angular/material/table'
import { MatIconModule } from '@angular/material/icon'
import { By } from '@angular/platform-browser'
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http'
import { DataService } from 'src/app/services/data/data.service'
import { MatSnackBar } from '@angular/material/snack-bar'
import { Overlay } from '@angular/cdk/overlay'

describe('RuleComponent', () => {
  let component: RuleComponent
  let fixture: ComponentFixture<RuleComponent>

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    declarations: [RuleComponent],
    imports: [MatTableModule, MatIconModule],
    providers: [DataService, MatSnackBar, Overlay, provideHttpClient(withInterceptorsFromDi())]
}).compileComponents()
  })

  beforeEach(() => {
    fixture = TestBed.createComponent(RuleComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
    component.dataSource = [
      {
        Name: 'varchar',
        Type: 'global_datatype_change',
        ObjectType: 'column',
        AssociatedObjects: 'all tables',
        Enabled: true,
      },
    ]
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })

  it('should render ADD rule button', () => {
    fixture.detectChanges()
    let button = fixture.debugElement.query(By.css('.add-rule-text'))
    expect(button.nativeElement.textContent).toEqual('ADD')
  })

  it('should render empty message when there is no data in the table', () => {
    component.dataSource = []
    fixture.detectChanges()
    let message = fixture.debugElement.query(By.css('.empty-message-content'))
    expect(message.nativeElement.textContent).toEqual(
      'No rules added yet. Add one by clicking "Add rule".'
    )
  })
  it('should always render table header even if there is no data for table', () => {
    fixture.detectChanges()
    let table = fixture.debugElement.query(By.css('.mat-column-name'))
    expect(table.nativeElement.textContent).toEqual('Rule name')
  })

  it('should render table data correctly', () => {
    fixture.detectChanges()

    let tableRows = fixture.nativeElement.querySelectorAll('tr')
    expect(tableRows.length).toBe(2)

    let row1 = tableRows[1]
    expect(row1.cells[0].innerText).toBe('1')
    expect(row1.cells[1].innerText).toBe('varchar')
    expect(row1.cells[2].innerText).toBe('Global Datatype Change')
    expect(row1.cells[3].innerText).toBe('column')
    expect(row1.cells[4].innerText).toBe('all tables')
    expect(row1.cells[5].innerText).toBe('Yes')
  })
})
