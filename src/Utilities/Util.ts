/* eslint-disable @typescript-eslint/no-extraneous-class */
export class Util {
    public static formatDate(dateFormat: Intl.DateTimeFormat, date: Date | number = new Date()): string {
        const data = dateFormat.formatToParts(date);
        return "<year>-<month>-<day>"
            .replace(/<year>/g, data.find(d => d.type === "year")!.value)
            .replace(/<month>/g, data.find(d => d.type === "month")!.value)
            .replace(/<day>/g, data.find(d => d.type === "day")!.value);
    }
}
