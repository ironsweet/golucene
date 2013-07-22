golucene
========

Go port of Apache Lucene but focus on IndexSearcher

Progress
--------
6500/40000 LoC, 33 Weeks Left

License
-------

Until further notice, all the code and document are licensed under APL 2.0.

Develop Guidelines
------------------

- NewXXX method can return value or pointer. Value is for simple data object, while pointer is for more complex action object or interface when you want to protect its internal fields.
- XXXReader and XXXDirectory must be pointers as they make changes to their underlying data structure.
- Parameter in pointer form indicates mutability.
- Use interface/inheritance when children don't add new capabilities.
- Use embeded interface/struct when children add new capabilities and explicitly used.
- Prefer value internally unless (a) optional field; or (b) obvious readonly fields.
- "not implemented yet" means in scope but not yet translated from Java.
- "not supported yet" means not in scope.
