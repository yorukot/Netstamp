export interface CurrentUser {
	name: string;
	username: string;
	email: string;
	role: string;
	gravatarUrl: string;
}

export const currentUser: CurrentUser = {
	name: "Elvis Mao",
	username: "elvis",
	email: "elvis@netstamp.dev",
	role: "Admin",
	gravatarUrl: "https://gravatar.com/avatar/f5a410169cdb93933383e6e54ac33b82e417fe84ffc4ed742adafd800cc07ab2?s=160&d=identicon"
};
