#include <QTextStream>
#include <QString>

int main() {
    QTextStream out(stdout);
    out << QString("Hello World");
}
