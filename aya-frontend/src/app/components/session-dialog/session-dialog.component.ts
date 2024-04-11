import { Component, Inject, inject, Input, OnInit } from '@angular/core';
import { MatFormField } from '@angular/material/form-field';
import {
  AbstractControl,
  FormArray,
  FormBuilder,
  FormControl,
  FormGroup,
  FormRecord,
  FormsModule,
  ReactiveFormsModule,
  ValidationErrors,
  ValidatorFn,
} from '@angular/forms';
import { MatInput, MatInputModule } from '@angular/material/input';
import { MatButton, MatIconButton } from '@angular/material/button';
import {
  MAT_DIALOG_DATA,
  MatDialogActions,
  MatDialogClose,
  MatDialogContent,
  MatDialogRef,
  MatDialogTitle,
} from '@angular/material/dialog';
import { MatOption, MatSelect } from '@angular/material/select';
import { MatIcon } from '@angular/material/icon';
import { MatToolbar } from '@angular/material/toolbar';
import { Validators } from '@angular/forms';
import { SessionDialogInfo } from '../../interfaces/session';

const sessionValidator: ValidatorFn = (
  control: AbstractControl
): ValidationErrors | null => {
  const resourceType = control.get('resourceType')?.value;
  switch (resourceType) {
    case 'discord':
      const discordChannelId = control.get('discordChannelId')?.value;
      const discordGuildId = control.get('discordGuildId')?.value;
      if (!discordChannelId) {
        return {
          missingDiscordChannelId: true,
        };
      } else if (!discordGuildId) {
        return {
          missingDiscordGuildId: true,
        };
      } else {
        return null;
      }
    case 'youtube':
      const youtubeChannelId = control.get('youtubeChannelId')?.value;
      if (!youtubeChannelId) {
        return {
          missingYoutubeChannelId: true,
        };
      } else {
        return null;
      }
    default:
      return {
        unknownResourceType: true,
      };
  }
};

@Component({
  selector: 'app-session-dialog',
  standalone: true,
  imports: [
    MatFormField,
    FormsModule,
    MatInput,
    MatButton,
    MatDialogTitle,
    MatDialogContent,
    MatDialogActions,
    MatDialogClose,
    ReactiveFormsModule,
    MatSelect,
    MatOption,
    MatIconButton,
    MatIcon,
    MatFormField,
    MatInputModule,
    MatToolbar,
  ],
  templateUrl: 'session-dialog.component.html',
  styleUrl: 'session-dialog.component.css',
})
export class SessionDialogComponent implements OnInit {
  constructor(
    public dialogRef: MatDialogRef<SessionDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: SessionDialogInfo
  ) {}

  private readonly formBuilder = inject(FormBuilder);

  sessionFormGroup = new FormGroup({
    resources: new FormArray([], {
      validators: [Validators.maxLength(3), Validators.minLength(0)],
    }),
  });

  get resources() {
    return this.sessionFormGroup.get('resources') as FormArray;
  }

  pushNewForm() {
    this.resources.push(
      new FormGroup(
        {
          resourceType: new FormControl('discord'),
          discordChannelId: new FormControl(''),
          discordGuildId: new FormControl(''),
          youtubeChannelId: new FormControl(''),
        },
        {
          validators: [sessionValidator],
        }
      )
    );
  }

  onNoClick() {
    this.dialogRef.close();
  }

  deleteForm(id: number) {
    if (id < 0 || id >= this.resources.length) {
      return;
    }
    this.resources.removeAt(id);
  }

  getAsFormGroup(form: AbstractControl) {
    return form as FormGroup;
  }

  getAsFormRecord(form: AbstractControl) {
    return form as FormRecord;
  }

  protected readonly String = String;
  protected readonly Validators = Validators;

  ngOnInit(): void {}
}
