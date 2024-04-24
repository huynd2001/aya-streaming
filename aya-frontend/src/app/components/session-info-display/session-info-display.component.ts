import { Component, inject, Input, OnInit } from "@angular/core";
import {
  ResourceInfo,
  SessionInfo,
  DisplaySessionInfo,
} from "../../interfaces/session";
import {
  MatAccordion,
  MatExpansionPanel,
  MatExpansionPanelDescription,
  MatExpansionPanelTitle,
} from "@angular/material/expansion";
import { MatIcon, MatIconRegistry } from "@angular/material/icon";
import { DomSanitizer } from "@angular/platform-browser";

@Component({
  selector: "app-session-info-display",
  standalone: true,
  imports: [
    MatExpansionPanel,
    MatExpansionPanelTitle,
    MatExpansionPanelDescription,
    MatIcon,
    MatAccordion,
  ],
  templateUrl: "./session-info-display.component.html",
  styleUrl: "./session-info-display.component.css",
})
export class SessionInfoDisplayComponent implements OnInit {
  @Input() displaySessionInfo: DisplaySessionInfo | undefined;
  public resources: ResourceInfo[] = [];

  private readonly matIconRegistry = inject(MatIconRegistry);
  private readonly domSanitizer = inject(DomSanitizer);

  validateResources(resources: any): resources is ResourceInfo[] {
    return Array.isArray(resources);
  }

  setResources() {
    if (this.displaySessionInfo === undefined) {
      this.resources = [];
      return;
    }
    const resourceParsed = JSON.parse(
      this.displaySessionInfo.session_info.Resources,
    );
    if (this.validateResources(resourceParsed)) {
      this.resources = resourceParsed.map((resource) => {
        return {
          resourceType: resource?.resourceType || "discord",
          resourceInfo: {
            discordChannelId: resource?.resourceInfo?.discordChannelId,
            discordGuildId: resource?.resourceInfo?.discordGuildId,
            youtubeChannelId: resource?.resourceInfo?.youtubeChannelId,
          },
        };
      });
      return;
    } else {
      this.resources = [];
    }
  }

  protected readonly JSON = JSON;

  ngOnInit(): void {
    this.matIconRegistry.addSvgIcon(
      `discord_logo`,
      this.domSanitizer.bypassSecurityTrustResourceUrl("/discord.svg"),
    );
    this.matIconRegistry.addSvgIcon(
      `youtube_logo`,
      this.domSanitizer.bypassSecurityTrustResourceUrl("/youtube.svg"),
    );
    this.setResources();
  }
}
