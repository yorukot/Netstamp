import netstampFavicon from "@netstamp/brand/assets/netstamp-favicon.svg";
import { Helmet } from "react-helmet-async";
import { AppRouter } from "./routes/AppRouter";
import { ThemeProvider } from "./shared/theme/ThemeProvider";

export default function App() {
	return (
		<ThemeProvider>
			<Helmet>
				<link rel="icon" type="image/svg+xml" href={netstampFavicon} />
			</Helmet>
			<AppRouter />
		</ThemeProvider>
	);
}
