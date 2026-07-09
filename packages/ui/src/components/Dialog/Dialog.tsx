import * as AlertDialogPrimitive from "@radix-ui/react-alert-dialog";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { forwardRef, type ComponentPropsWithoutRef, type ComponentRef } from "react";
import styles from "./Dialog.module.css";

function classNames(...values: Array<string | false | null | undefined>) {
	return values.filter(Boolean).join(" ");
}

export const DialogRoot = DialogPrimitive.Root;
export const DialogTrigger = DialogPrimitive.Trigger;
export const DialogPortal = DialogPrimitive.Portal;
export const DialogOverlay = DialogPrimitive.Overlay;
export const DialogContent = DialogPrimitive.Content;
export const DialogTitle = DialogPrimitive.Title;
export const DialogDescription = DialogPrimitive.Description;
export const DialogClose = DialogPrimitive.Close;

export type DialogRootProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Root>;
export type DialogTriggerProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Trigger>;
export type DialogPortalProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Portal>;
export type DialogOverlayProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>;
export type DialogContentProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Content>;
export type DialogTitleProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Title>;
export type DialogDescriptionProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Description>;
export type DialogCloseProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Close>;

export const AlertDialogRoot = AlertDialogPrimitive.Root;
export const AlertDialogTrigger = AlertDialogPrimitive.Trigger;
export const AlertDialogPortal = AlertDialogPrimitive.Portal;
export const AlertDialogOverlay = AlertDialogPrimitive.Overlay;
export const AlertDialogContent = AlertDialogPrimitive.Content;
export const AlertDialogTitle = AlertDialogPrimitive.Title;
export const AlertDialogDescription = AlertDialogPrimitive.Description;
export const AlertDialogAction = AlertDialogPrimitive.Action;
export const AlertDialogCancel = AlertDialogPrimitive.Cancel;

export type AlertDialogRootProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Root>;
export type AlertDialogTriggerProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Trigger>;
export type AlertDialogPortalProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Portal>;
export type AlertDialogOverlayProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Overlay>;
export type AlertDialogContentProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Content>;
export type AlertDialogTitleProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Title>;
export type AlertDialogDescriptionProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Description>;
export type AlertDialogActionProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Action>;
export type AlertDialogCancelProps = ComponentPropsWithoutRef<typeof AlertDialogPrimitive.Cancel>;

export const DropdownMenuRoot = DropdownMenuPrimitive.Root;
export const DropdownMenuTrigger = DropdownMenuPrimitive.Trigger;
export const DropdownMenuPortal = DropdownMenuPrimitive.Portal;
export const DropdownMenuContent = forwardRef<ComponentRef<typeof DropdownMenuPrimitive.Content>, DropdownMenuContentProps>(function DropdownMenuContent({ className, ...props }, ref) {
	return <DropdownMenuPrimitive.Content ref={ref} className={classNames(styles.dropdownMenuContent, className)} {...props} />;
});
export const DropdownMenuItem = DropdownMenuPrimitive.Item;
export const DropdownMenuCheckboxItem = DropdownMenuPrimitive.CheckboxItem;

export type DropdownMenuRootProps = ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Root>;
export type DropdownMenuTriggerProps = ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Trigger>;
export type DropdownMenuPortalProps = ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Portal>;
export type DropdownMenuContentProps = ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Content>;
export type DropdownMenuItemProps = ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Item>;
export type DropdownMenuCheckboxItemProps = ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.CheckboxItem>;

export const PopoverRoot = PopoverPrimitive.Root;
export const PopoverTrigger = PopoverPrimitive.Trigger;
export const PopoverAnchor = PopoverPrimitive.Anchor;
export const PopoverPortal = PopoverPrimitive.Portal;
export const PopoverContent = forwardRef<ComponentRef<typeof PopoverPrimitive.Content>, PopoverContentProps>(function PopoverContent({ className, ...props }, ref) {
	return <PopoverPrimitive.Content ref={ref} className={classNames(styles.popoverContent, className)} {...props} />;
});
export const PopoverClose = PopoverPrimitive.Close;

export type PopoverRootProps = ComponentPropsWithoutRef<typeof PopoverPrimitive.Root>;
export type PopoverTriggerProps = ComponentPropsWithoutRef<typeof PopoverPrimitive.Trigger>;
export type PopoverAnchorProps = ComponentPropsWithoutRef<typeof PopoverPrimitive.Anchor>;
export type PopoverPortalProps = ComponentPropsWithoutRef<typeof PopoverPrimitive.Portal>;
export type PopoverContentProps = ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>;
export type PopoverCloseProps = ComponentPropsWithoutRef<typeof PopoverPrimitive.Close>;
