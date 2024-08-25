import React, { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from "@/components/ui/command";
import type { LynxCommandGroup } from "@/lib/CommandMenuContext";
import URLS from "@/lib/urls";

interface CommandMenuProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  customCommands?: LynxCommandGroup[];
}

const CommandMenu: React.FC<CommandMenuProps> = ({
  open,
  onOpenChange,
  customCommands,
}) => {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const runCommand = (command: () => void) => {
    onOpenChange(false);
    command();
  };
  const navigateHome = useCallback(
    () => runCommand(() => navigate(URLS.HOME)),
    [navigate],
  );
  const navigateSettings = useCallback(
    () => runCommand(() => navigate(URLS.SETTINGS)),
    [navigate],
  );
  const navigateFeeds = useCallback(
    () => runCommand(() => navigate(URLS.FEEDS)),
    [navigate],
  );
  const navigateCookies = useCallback(
    () => runCommand(() => navigate(URLS.COOKIES)),
    [navigate],
  );
  const navigateApiKeys = useCallback(
    () => runCommand(() => navigate(URLS.API_KEYS)),
    [navigate],
  );
  const navigateAddLink = useCallback(
    () => runCommand(() => navigate(URLS.ADD_LINK)),
    [navigate],
  );

  return (
    <CommandDialog open={open} onOpenChange={onOpenChange}>
      <CommandInput
        value={search}
        onValueChange={setSearch}
        placeholder="Type a command or search..."
      />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        {(customCommands || []).map((group) => (
          <CommandGroup key={group.display} heading={group.display}>
            {group.items.map((item) => (
              <CommandItem
                onSelect={item.onSelect}
                key={group.display + item.display}
              />
            ))}
          </CommandGroup>
        ))}
        <CommandGroup heading="Pages">
          <CommandItem onSelect={navigateHome}>Home</CommandItem>
          <CommandItem onSelect={navigateSettings}>Settings</CommandItem>
          <CommandItem onSelect={navigateApiKeys}>API Keys</CommandItem>
          <CommandItem onSelect={navigateCookies}>Cookies</CommandItem>
          <CommandItem onSelect={navigateFeeds}>Feeds</CommandItem>
          <CommandItem onSelect={navigateAddLink}>Add Link</CommandItem>
        </CommandGroup>
        {search !== "" && (
          <CommandGroup heading="Search">
            <CommandItem
              onSelect={() =>
                runCommand(() => navigate(URLS.HOME_WITH_SEARCH_STRING(search)))
              }
            >
              Search for "{search}"
            </CommandItem>
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
};

export default CommandMenu;
