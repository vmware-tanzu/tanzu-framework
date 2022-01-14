import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TanzuUiApiLibComponent } from './tanzu-ui-api-lib.component';

describe('TanzuUiApiLibComponent', () => {
  let component: TanzuUiApiLibComponent;
  let fixture: ComponentFixture<TanzuUiApiLibComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TanzuUiApiLibComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(TanzuUiApiLibComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
