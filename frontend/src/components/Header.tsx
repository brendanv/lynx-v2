import { useCallback } from "react";
import { CirclePlus, CircleUser, Menu, Search } from "lucide-react";
import { Link, useNavigate, useLocation } from "react-router-dom";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { usePocketBase } from "@/hooks/usePocketBase";
import URLS from "@/lib/urls";
import { useCommandMenu } from "@/lib/CommandMenuContext";
import { ThemeToggle } from "@/components/ThemeToggle";
import { useTheme } from "@/components/ThemeProvider";

const Header = () => {
  const { pb } = usePocketBase();
  const { setCommandMenuOpen } = useCommandMenu();
  const location = useLocation();
  const navigate = useNavigate();
  const handleLogout = useCallback(() => {
    pb.authStore.clear();
    navigate(URLS.LOGIN);
  }, [pb.authStore]);
  const openCommandMenu = useCallback(() => {
    setCommandMenuOpen(true);
  }, [setCommandMenuOpen]);
  const { appliedTheme } = useTheme();

  const menuBarOtherClassName =
    "text-muted-foreground transition-colors hover:text-foreground";
  const menuBarCurrentClassName =
    "text-foreground transition-colors hover:text-foreground";

  const sideBarOtherClassName = "text-muted-foreground hover:text-foreground";
  const sideBarCurrentClassName = "text-foreground hover:text-foreground";

  return (
    <header className="sticky top-0 flex h-16 items-center gap-4 border-b bg-background px-4 md:px-6 z-50">
      <nav className="hidden flex-col gap-6 text-lg font-medium md:flex md:flex-row md:items-center md:gap-5 md:text-sm lg:gap-6">
        <Link
          to={URLS.HOME}
          className="flex items-center gap-2 text-lg font-semibold md:text-base"
        >
          <div className="h-12 w-12">
            {appliedTheme == "dark" ? (
              <img src="/img/lynx_light.svg" />
            ) : (
              <img src="/img/lynx_dark.svg" />
            )}
          </div>
          <span className="sr-only">Lynx</span>
        </Link>
        <Link
          to={URLS.HOME}
          className={
            location.pathname === URLS.HOME
              ? menuBarCurrentClassName
              : menuBarOtherClassName
          }
        >
          Home
        </Link>
        <Link
          to={URLS.NOTES}
          className={
            location.pathname === URLS.NOTES
              ? menuBarCurrentClassName
              : menuBarOtherClassName
          }
        >
          Highlights
        </Link>
        <Link
          to={URLS.TAGS}
          className={
            location.pathname === URLS.TAGS
              ? menuBarCurrentClassName
              : menuBarOtherClassName
          }
        >
          Tags
        </Link>
        <Link
          to={URLS.FEEDS}
          className={
            location.pathname === URLS.FEEDS
              ? menuBarCurrentClassName
              : menuBarOtherClassName
          }
        >
          Feeds
        </Link>
        <Link
          to={URLS.SETTINGS}
          className={
            location.pathname === URLS.SETTINGS
              ? menuBarCurrentClassName
              : menuBarOtherClassName
          }
        >
          Settings
        </Link>
      </nav>
      <Sheet>
        <SheetTrigger asChild>
          <Button variant="outline" size="icon" className="shrink-0 md:hidden">
            <Menu className="h-5 w-5" />
            <span className="sr-only">Toggle navigation menu</span>
          </Button>
        </SheetTrigger>
        <SheetContent side="left">
          <nav className="grid gap-6 text-lg font-medium">
            <Link
              to={URLS.HOME}
              className="flex items-center gap-2 text-lg font-semibold"
            >
              <div className="h-36 w-36">
                {appliedTheme == "dark" ? (
                  <img src="/img/lynx_light.svg" />
                ) : (
                  <img src="/img/lynx_dark.svg" />
                )}
              </div>
              <span className="sr-only">Lynx</span>
            </Link>
            <Link
              to={URLS.HOME}
              className={
                location.pathname === URLS.HOME
                  ? sideBarCurrentClassName
                  : sideBarOtherClassName
              }
            >
              Home
            </Link>
            <Link
              to={URLS.NOTES}
              className={
                location.pathname === URLS.NOTES
                  ? sideBarCurrentClassName
                  : sideBarOtherClassName
              }
            >
              Highlights
            </Link>
            <Link
              to={URLS.TAGS}
              className={
                location.pathname === URLS.TAGS
                  ? sideBarCurrentClassName
                  : sideBarOtherClassName
              }
            >
              Tags
            </Link>
            <Link
              to={URLS.FEEDS}
              className={
                location.pathname === URLS.FEEDS
                  ? sideBarCurrentClassName
                  : sideBarOtherClassName
              }
            >
              Feeds
            </Link>
            <Link
              to={URLS.SETTINGS}
              className={
                location.pathname === URLS.SETTINGS
                  ? sideBarCurrentClassName
                  : sideBarOtherClassName
              }
            >
              Settings
            </Link>
            <ThemeToggle />
          </nav>
        </SheetContent>
      </Sheet>
      <div className="flex w-full items-center gap-4 md:ml-auto md:gap-2 lg:gap-4">
        <form className="ml-auto flex-1 sm:flex-initial">
          <div className="relative">
            <Search className="absolute left-2.5 top-3 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search Lynx..."
              className="pl-8 sm:w-[300px] md:w-[200px] lg:w-[300px]"
              onSelect={openCommandMenu}
            />
            <kbd className="pointer-events-none absolute h-4 right-2.5 top-3 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
              ⌘K
            </kbd>
          </div>
        </form>
        <Link to={URLS.ADD_LINK}>
          <Button variant="ghost" size="icon" className="rounded-full">
            <CirclePlus className="h-5 w-5" />
          </Button>
        </Link>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="rounded-full">
              <CircleUser className="h-5 w-5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link to={URLS.SETTINGS}>Settings</Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={handleLogout}>Logout</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <span className="hidden md:flex">
          <ThemeToggle />
        </span>
      </div>
    </header>
  );
};

export default Header;
