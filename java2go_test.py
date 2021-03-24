import unittest
# import os

# Import all the functionality from the main file
import java2go

class TestSimpleJavaFile(unittest.TestCase):


    def test_conversion(self):
        self.maxDiff = None

        simpleJava = ''
        simpleGolang = ''

        with open('testsnippets/simple.java', 'r') as javaFile:
            simpleJava = javaFile.read()

        with open('testsnippets/simple.go', 'r') as goFile:
            simpleGolang = goFile.read()

        with open('generated_code.go', 'w') as outfile:
            outfile.write(java2go.convert(simpleJava))

        self.assertEqual(java2go.convert(simpleJava), simpleGolang)

if __name__ == '__main__':
    unittest.main()
