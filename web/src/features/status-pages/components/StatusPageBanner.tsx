import { useState } from "react";

type StatusPageBannerProps = {
	className?: string;
	src?: string;
};

const BannerImage = ({ className, src }: { className?: string; src: string }) => {
	const [failed, setFailed] = useState(false);

	if (failed) return null;

	return <img className={className} src={src} alt="" onError={() => setFailed(true)} />;
};

export const StatusPageBanner = ({ className, src }: StatusPageBannerProps) => {
	const normalizedSrc = src?.trim();

	if (!normalizedSrc) return null;

	return <BannerImage key={normalizedSrc} className={className} src={normalizedSrc} />;
};
