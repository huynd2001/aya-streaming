import { Component, Inject, inject, OnInit } from '@angular/core';
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
  MatDialogModule,
  MatDialogRef,
  MatDialogTitle,
} from '@angular/material/dialog';
import { MatOption, MatSelect } from '@angular/material/select';
import { MatIcon } from '@angular/material/icon';
import { MatToolbar } from '@angular/material/toolbar';
import { Validators } from '@angular/forms';
import { ResourceInfo, SessionDialogInfo } from '../../interfaces/session';

function sessionValidator(): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const resourceType = control.get('resourceType')?.value;
    switch (resourceType) {
      case 'discord':
        const discordChannelId = control
          .get('resourceInfo')
          ?.get('discordChannelId')?.value;
        const discordGuildId = control
          .get('resourceInfo')
          ?.get('discordGuildId')?.value;
        let validationError = {};
        if (!discordChannelId) {
          validationError = {
            missingDiscordChannelId: true,
            ...validationError,
          };
        }
        if (!discordGuildId) {
          validationError = {
            missingDiscordGuildId: true,
            ...validationError,
          };
        }
        if (Object.keys(validationError).length === 0) {
          return null;
        } else {
          return validationError;
        }

      case 'youtube':
        const youtubeChannelId = control
          .get('resourceInfo')
          ?.get('youtubeChannelId')?.value;
        if (!youtubeChannelId) {
          return {
            missingYoutubeChannelId: true,
          };
        } else {
          return null;
        }

      case 'twitch':
        const twitchChannelName = control
          .get('resourceInfo')
          ?.get('twitchChannelName')?.value;
        if (!twitchChannelName) {
          return {
            missingTwitchChannelName: true,
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
}

@Component({
  selector: 'app-session-dialog',
  standalone: true,
  imports: [
    MatFormField,
    FormsModule,
    MatInput,
    MatButton,
    MatDialogModule,
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
    @Inject(MAT_DIALOG_DATA) public data: SessionDialogInfo | undefined,
  ) {}

  sessionFormGroup = new FormGroup({
    resources: new FormArray([], {
      validators: [Validators.maxLength(3), Validators.minLength(0)],
    }),
  });

  validate(inputData: any) {
    return (
      inputData == undefined ||
      ((inputData.id === undefined || typeof inputData.id === 'number') &&
        Array.isArray(inputData.resources))
    );
  }

  populateForm() {
    if (this.data) {
      for (let resource of this.data.resources) {
        this.resources.push(
          new FormGroup(
            {
              resourceType: new FormControl(
                resource?.resourceType || 'discord',
              ),
              resourceInfo: new FormGroup({
                discordChannelId: new FormControl(
                  resource?.resourceInfo?.discordChannelId || '',
                ),
                discordGuildId: new FormControl(
                  resource?.resourceInfo?.discordGuildId || '',
                ),
                youtubeChannelId: new FormControl(
                  resource?.resourceInfo?.youtubeChannelId || '',
                ),
                twitchChannelName: new FormControl(
                  resource?.resourceInfo?.twitchChannelName || '',
                ),
              }),
            },
            {
              validators: [sessionValidator()],
            },
          ),
        );
      }
    }
  }

  ngOnInit(): void {
    if (this.validate(this.data)) {
      this.populateForm();
    } else {
    }
  }

  private readonly formBuilder = inject(FormBuilder);

  get resources() {
    return this.sessionFormGroup.get('resources') as FormArray;
  }

  pushNewForm() {
    this.resources.push(
      new FormGroup(
        {
          resourceType: new FormControl('discord'),
          resourceInfo: new FormGroup({
            discordChannelId: new FormControl(''),
            discordGuildId: new FormControl(''),
            youtubeChannelId: new FormControl(''),
            twitchChannelName: new FormControl(''),
          }),
        },
        {
          validators: [sessionValidator()],
        },
      ),
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

  retrieveValue() {
    let dialogInfo = this.sessionFormGroup.value as SessionDialogInfo;
    let retDialog: SessionDialogInfo = {
      id: this.data?.id,
      resources: [],
    };
    for (let resource of dialogInfo.resources) {
      let newResource: ResourceInfo = {
        resourceType: resource.resourceType,
        resourceInfo: {},
      };
      switch (resource.resourceType) {
        case 'discord':
          newResource.resourceInfo.discordGuildId =
            resource.resourceInfo.discordGuildId;
          newResource.resourceInfo.discordChannelId =
            resource.resourceInfo.discordChannelId;
          break;
        case 'youtube':
          newResource.resourceInfo.youtubeChannelId =
            resource.resourceInfo.youtubeChannelId;
          break;
        case 'twitch':
          newResource.resourceInfo.twitchChannelName =
            resource.resourceInfo.twitchChannelName;
          break;
        default:
          break;
      }
      retDialog.resources.push(newResource);
    }
    return retDialog;
  }
}
